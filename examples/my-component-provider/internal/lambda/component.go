// Package lambda contains a component that creates a lambda function, with some presets
package lambda

import (
	"fmt"

	cloudwatch "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/cloudwatch"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/core"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DefaultLambdaConfig returns a LambdaConfig populated with the same
func DefaultLambdaConfig() LambdaConfig {
	return LambdaConfig{
		Timeout:          60 * 15,
		MemorySize:       8000,
		MaxConcurrency:   1,
		RetryLimit:       3,
		LogRetentionDays: 365,
	}
}

// LambdaConfig mirrors the Python LambdaConfig dataclass.
type LambdaConfig struct {
	Timeout          int
	MemorySize       int
	MaxConcurrency   int
	RetryLimit       int
	LogRetentionDays int
	AlertsEmail      *string
	KmsKeyId         *string // Optional KMS Key ID for log encryption
}

// withDefaults applies Python-equivalent defaults to any zero valued fields.
func (c LambdaConfig) withDefaults() LambdaConfig {
	if c.Timeout == 0 {
		c.Timeout = 60 * 15
	}
	if c.MemorySize == 0 {
		c.MemorySize = 8000
	}
	if c.MaxConcurrency == 0 {
		c.MaxConcurrency = 1
	}
	if c.RetryLimit == 0 {
		c.RetryLimit = 3
	}
	if c.LogRetentionDays == 0 {
		c.LogRetentionDays = 365
	}
	return c
}

// LambdaComponentArgs defines inputs for NewLambdaComponent.
type LambdaComponentArgs struct {
	LambdaConfig       LambdaConfig
	EcrRepo            *ecr.Repository
	AdditionalEnvs     map[string]string
	S3TriggerBucket    pulumi.StringInput
	S3BucketArns       []string
	KmsKeyArns         []string
	SecretsManagerArns []string
	Provider           *aws.Provider
}

// LambdaComponent creates a basic Lambda function from an ECR image.
type LambdaComponent struct {
	pulumi.ResourceState

	FunctionArn pulumi.StringOutput
	RoleArn     pulumi.StringOutput
	Function    *awslambda.Function
}

// NewLambdaComponent provisions a Lambda function with a role and optional
// permissions. The implementation is intentionally simplified compared to the
// Python version but keeps the same overall behaviour.
func NewLambdaComponent(ctx *pulumi.Context, name string, args *LambdaComponentArgs, opts ...pulumi.ResourceOption) (*LambdaComponent, error) {
	if args == nil {
		args = &LambdaComponentArgs{}
	}

	cfg := args.LambdaConfig.withDefaults()

	comp := &LambdaComponent{}
	err := ctx.RegisterComponentResource("gs:lambda:LambdaComponent", name, comp, opts...)
	if err != nil {
		return nil, err
	}

	options := append(opts, pulumi.Parent(comp))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	id, err := aws.GetCallerIdentity(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Base policy statements for log access.
	statements := []iam.GetPolicyDocumentStatement{
		{
			Effect: pulumi.StringRef("Allow"),
			Actions: []string{
				"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents",
			},
			Resources: []string{"arn:aws:logs:*:*:*"},
		},
	}

	if len(args.S3BucketArns) > 0 {
		statements = append(statements, iam.GetPolicyDocumentStatement{
			Effect:    pulumi.StringRef("Allow"),
			Actions:   []string{"s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"},
			Resources: args.S3BucketArns,
		})
	}

	if len(args.KmsKeyArns) > 0 {
		statements = append(statements, iam.GetPolicyDocumentStatement{
			Effect:    pulumi.StringRef("Allow"),
			Actions:   []string{"kms:Decrypt", "kms:Encrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey"},
			Resources: args.KmsKeyArns,
		})
	}

	if len(args.SecretsManagerArns) > 0 {
		statements = append(statements, iam.GetPolicyDocumentStatement{
			Effect:    pulumi.StringRef("Allow"),
			Actions:   []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"},
			Resources: args.SecretsManagerArns,
		})
	}

	policyDoc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: statements,
	})
	if err != nil {
		return nil, err
	}

	role, err := iam.NewRole(ctx, fmt.Sprintf("%s-role", name), &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}`),
	}, options...)
	if err != nil {
		return nil, err
	}

	policy, err := iam.NewPolicy(ctx, fmt.Sprintf("%s-policy", name), &iam.PolicyArgs{
		Policy: pulumi.String(policyDoc.Json),
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-attach", name), &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: policy.Arn,
	}, options...)
	if err != nil {
		return nil, err
	}

	//
	// image := args.EcrRepo.RepositoryUrl.ApplyT(func(u string) string {
	// 	return fmt.Sprintf("%s:latest", u)
	// }).(pulumi.StringOutput)

	// image := args.EcrRepo.Rep.ApplyT(func(u string) string {
	// 	return fmt.Sprintf("%s:latest", u)
	// }).(pulumi.StringOutput)

	image := args.EcrRepo.Name.ApplyT(func(repoName string) (string, error) {
		image, err := ecr.GetImage(ctx, &ecr.GetImageArgs{
			RepositoryName: repoName,
			MostRecent:     pulumi.BoolRef(true),
		}, nil)
		if err != nil {
			return "", err
		}
		return image.Id, nil
	}).(pulumi.StringOutput)

	//
	// //have to use ApplyT to get repo name for pulumi
	// image, err := args.EcrRepo.Name.ApplyT(func(repo string) {
	// 	ecr.GetImage(ctx, &ecr.GetImageArgs{
	// 		RepositoryName: pulumi.String(repo),
	// 		MostRecent: true,
	// 	}, nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
	//

	// image, err := ecr.GetImage(ctx, &ecr.GetImageArgs{
	// 		RepositoryName: args.EcrRepo.Name,
	// 		ImageTag:       pulumi.StringRef("latest"),
	// 	}, nil)
	//
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//

	envs := pulumi.StringMap{
		"ACCOUNT_ID": pulumi.String(id.AccountId),
	}
	for k, v := range args.AdditionalEnvs {
		envs[k] = pulumi.String(v)
	}

	fn, err := awslambda.NewFunction(ctx, fmt.Sprintf("%s-func", name), &awslambda.FunctionArgs{
		PackageType: pulumi.String("Image"),
		ImageUri:    image,
		Role:        role.Arn,
		Timeout:     pulumi.Int(cfg.Timeout),
		MemorySize:  pulumi.Int(cfg.MemorySize),
		Environment: awslambda.FunctionEnvironmentArgs{Variables: envs},
		Tags:        core.GenerateBaseTags(ctx),
	}, options...)
	if err != nil {
		return nil, err
	}

	comp.Function = fn

	if cfg.AlertsEmail != nil && *cfg.AlertsEmail != "" {
		fmt.Println("Would have created cloudwatch component")

		_, err = cloudwatch.NewLambdaCloudwatchComponent(ctx, fmt.Sprintf("%s-cw", name), &cloudwatch.LambdaCloudwatchConfig{
			Function:         fn,
			LogRetentionDays: cfg.LogRetentionDays,
			AlertsEmail:      *cfg.AlertsEmail,
			RetryLimit:       cfg.RetryLimit,
			// KmsKeyId:         []string{args.KmsKeyArns},
			KmsKeyId: cfg.KmsKeyId,
			Provider: args.Provider,
		}, options...)
		if err != nil {
			return nil, err
		}
	}

	if args.S3TriggerBucket != nil {
		srcArn := args.S3TriggerBucket.ToStringOutput().ApplyT(func(n string) string {
			return fmt.Sprintf("arn:aws:s3:::%s", n)
		}).(pulumi.StringOutput)
		_, err = awslambda.NewPermission(ctx, fmt.Sprintf("%s-s3perm", name), &awslambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  fn.Name,
			Principal: pulumi.String("s3.amazonaws.com"),
			SourceArn: srcArn.ToStringPtrOutput(),
		}, options...)
		if err != nil {
			return nil, err
		}

		_, err = s3.NewBucketNotification(ctx, fmt.Sprintf("%s-notify", name), &s3.BucketNotificationArgs{
			Bucket: args.S3TriggerBucket,
			LambdaFunctions: s3.BucketNotificationLambdaFunctionArray{
				s3.BucketNotificationLambdaFunctionArgs{
					LambdaFunctionArn: fn.Arn,
					Events:            pulumi.StringArray{pulumi.String("s3:ObjectCreated:*")},
				},
			},
		}, append(options, pulumi.DependsOn([]pulumi.Resource{fn}))...)
		if err != nil {
			return nil, err
		}
	}

	comp.FunctionArn = fn.Arn
	comp.RoleArn = role.Arn

	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"functionArn":  fn.Arn,
		"roleArn":      role.Arn,
		"functionName": fn.Name,
	}); err != nil {
		return nil, err
	}
	return comp, nil
}
