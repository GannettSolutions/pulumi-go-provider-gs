package s3

import (
	"fmt"

	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/core"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Args configure the S3BucketComponent.
type Args struct {
	KmsKeyArn pulumi.StringInput
	Provider  *aws.Provider
}

// S3BucketComponent creates an S3 bucket with encryption and a basic policy.
type S3BucketComponent struct {
	pulumi.ResourceState

	Bucket       *s3.Bucket
	BucketPolicy *s3.BucketPolicy
}

func NewS3BucketComponent(ctx *pulumi.Context, name string, args *Args, opts ...pulumi.ResourceOption) (*S3BucketComponent, error) {
	if args == nil || args.KmsKeyArn == nil {
		return nil, fmt.Errorf("KmsKeyArn is required")
	}

	comp := &S3BucketComponent{}
	if err := ctx.RegisterComponentResource("oneblood:s3:S3BucketComponent", name, comp, opts...); err != nil {
		return nil, err
	}

	options := append(opts, pulumi.Parent(comp))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	stack := ctx.Stack()
	project := ctx.Project()
	bucketName := fmt.Sprintf("%s-%s-%s", stack, project, name)

	bucket, err := s3.NewBucket(ctx, bucketName, &s3.BucketArgs{
		Bucket: pulumi.String(bucketName),
		Versioning: s3.BucketVersioningArgs{
			Enabled: pulumi.Bool(true),
		},
		ServerSideEncryptionConfiguration: s3.BucketServerSideEncryptionConfigurationArgs{
			Rule: s3.BucketServerSideEncryptionConfigurationRuleArgs{
				ApplyServerSideEncryptionByDefault: s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
					KmsMasterKeyId: args.KmsKeyArn,
					SseAlgorithm:   pulumi.String("aws:kms"),
				},
				BucketKeyEnabled: pulumi.Bool(true),
			},
		},
		Tags: core.GenerateBaseTags(ctx),
	}, options...)
	if err != nil {
		return nil, err
	}

	policyJSON := pulumi.All(bucket.ID()).ApplyT(func(vs []interface{}) (string, error) {
		b := string(vs[0].(pulumi.ID))
		return fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"EnforceSecureTransport","Effect":"Deny","Principal":"*","Action":"s3:*","Resource":["arn:aws:s3:::%s/*","arn:aws:s3:::%s"],"Condition":{"Bool":{"aws:SecureTransport":"false"}}}]}`, b, b), nil
	}).(pulumi.StringOutput)

	pol, err := s3.NewBucketPolicy(ctx, fmt.Sprintf("%s-policy", name), &s3.BucketPolicyArgs{
		Bucket: bucket.ID(),
		Policy: policyJSON,
	}, options...)
	if err != nil {
		return nil, err
	}

	comp.Bucket = bucket
	comp.BucketPolicy = pol

	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"bucket":       bucket.ID(),
		"bucketPolicy": pol.ID(),
	}); err != nil {
		return nil, err
	}

	return comp, nil
}
