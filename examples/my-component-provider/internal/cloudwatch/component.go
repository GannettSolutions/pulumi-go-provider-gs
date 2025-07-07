package cloudwatch

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// LogGroupArgs defines the arguments for creating a CloudWatch Log Group.
type LogGroupArgs struct {
	LogGroupName    pulumi.StringInput
	RetentionInDays pulumi.IntInput
	KmsKeyId        pulumi.StringInput
	Provider        pulumi.ProviderResource
}

// NewLogGroupComponent creates a CloudWatch Log Group.
func NewLogGroupComponent(ctx *pulumi.Context, name string, args *LogGroupArgs, opts ...pulumi.ResourceOption) (*cloudwatch.LogGroup, error) {
	if args == nil {
		args = &LogGroupArgs{}
	}

	options := append(opts)
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	logGroup, err := cloudwatch.NewLogGroup(ctx, name, &cloudwatch.LogGroupArgs{
		Name:            args.LogGroupName,
		RetentionInDays: args.RetentionInDays,
		KmsKeyId:        args.KmsKeyId,
	}, options...)
	if err != nil {
		return nil, err
	}

	return logGroup, nil
}

// LambdaCloudwatchArgs defines the arguments for the Lambda Cloudwatch Component.
type LambdaCloudwatchArgs struct {
	Function         pulumi.Resource
	LogRetentionDays int
	AlertsEmail      string
	RetryLimit       int
	KmsKeyId         pulumi.StringInput // New field for KMS Key ID
	Provider         pulumi.ProviderResource
}

// LambdaCloudwatchComponent represents a Cloudwatch component for Lambda.
