package cloudwatch

import (
	"fmt"

	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/core"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Args configure NewCloudwatchComponent.
type Args struct {
	Function         *awslambda.Function
	LogRetentionDays int
	AlertsEmail      string
	RetryLimit       int
	Provider         *aws.Provider
}

// CloudwatchComponent sets up logging and alerting for a Lambda function.
type CloudwatchComponent struct {
	pulumi.ResourceState

	LogGroupArn   pulumi.StringOutput
	AlertTopicArn pulumi.StringOutput
}

func NewCloudwatchComponent(ctx *pulumi.Context, name string, args *Args, opts ...pulumi.ResourceOption) (*CloudwatchComponent, error) {
	if args == nil || args.Function == nil {
		return nil, fmt.Errorf("Function must be provided")
	}
	comp := &CloudwatchComponent{}
	err := ctx.RegisterComponentResource("oneblood:cloudwatch:CloudwatchComponent", name, comp, opts...)
	if err != nil {
		return nil, err
	}
	options := append(opts, pulumi.Parent(comp))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	tags := core.GenerateBaseTags(ctx)

	logName := args.Function.Name.ApplyT(func(n string) string {
		return fmt.Sprintf("/aws/lambda/%s", n)
	}).(pulumi.StringOutput)
	lg, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-loggroup", name), &cloudwatch.LogGroupArgs{
		Name:            logName,
		RetentionInDays: pulumi.Int(args.LogRetentionDays),
		Tags:            tags,
	}, options...)
	if err != nil {
		return nil, err
	}

	topicName := args.Function.Name.ApplyT(func(n string) string {
		return fmt.Sprintf("%s-alerting", n)
	}).(pulumi.StringOutput)
	topic, err := sns.NewTopic(ctx, fmt.Sprintf("%s-topic", name), &sns.TopicArgs{
		Name: topicName,
		Tags: tags,
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = sns.NewTopicSubscription(ctx, fmt.Sprintf("%s-sub", name), &sns.TopicSubscriptionArgs{
		Topic:    topic.Arn,
		Protocol: pulumi.String("email"),
		Endpoint: pulumi.String(args.AlertsEmail),
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = cloudwatch.NewMetricAlarm(ctx, fmt.Sprintf("%s-error-alarm", name), &cloudwatch.MetricAlarmArgs{
		AlarmDescription: args.Function.Name.ApplyT(func(fn string) string {
			return fmt.Sprintf("Errors for %s", fn)
		}).(pulumi.StringOutput),
		ComparisonOperator: pulumi.String("GreaterThanOrEqualToThreshold"),
		EvaluationPeriods:  pulumi.Int(1),
		Threshold:          pulumi.Float64(float64(args.RetryLimit)),
		TreatMissingData:   pulumi.String("breaching"),
		AlarmActions:       pulumi.Array{topic.Arn},
		OkActions:          pulumi.Array{topic.Arn},
		MetricQueries: cloudwatch.MetricAlarmMetricQueryArray{
			cloudwatch.MetricAlarmMetricQueryArgs{
				Id:         pulumi.String("m1"),
				ReturnData: pulumi.Bool(true),
				Metric: cloudwatch.MetricAlarmMetricQueryMetricArgs{
					Dimensions: pulumi.StringMap{
						"FunctionName": args.Function.Name,
					},
					MetricName: pulumi.String("Errors"),
					Namespace:  pulumi.String("AWS/Lambda"),
					Period:     pulumi.Int(86400),
					Stat:       pulumi.String("Sum"),
					Unit:       pulumi.String("Count"),
				},
			},
		},
		Tags: tags,
	}, append(options, pulumi.DependsOn([]pulumi.Resource{topic}))...)
	if err != nil {
		return nil, err
	}

	comp.LogGroupArn = lg.Arn
	comp.AlertTopicArn = topic.Arn
	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"logGroupArn":   lg.Arn,
		"alertTopicArn": topic.Arn,
	}); err != nil {
		return nil, err
	}
	return comp, nil
}
