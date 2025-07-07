// Package cloudwatch contains a component that sets up logging and alerting for a Lambda function.
package cloudwatch

import (
	"fmt"

	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/core"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/scheduler"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DefaultLambdaCloudwatchConfig() LambdaCloudwatchConfig {
	return LambdaCloudwatchConfig{
		LogRetentionDays: 30,
		RetryLimit:       3,
		Timeout:          60,
	}
}

// LambdaCloudwatchConfig configure NewCloudwatchComponent.
type LambdaCloudwatchConfig struct {
	Function         *awslambda.Function
	LogRetentionDays int
	AlertsEmail      string
	RetryLimit       int
	Timeout          int
	KmsKeyId         *string
	Provider         *aws.Provider
}

// LambdaCloudwatchComponent sets up logging and alerting for a Lambda function.
type LambdaCloudwatchComponent struct {
	pulumi.ResourceState

	LogGroupArn   pulumi.StringOutput
	AlertTopicArn pulumi.StringOutput
}

func NewLambdaCloudwatchComponent(ctx *pulumi.Context, name string, args *LambdaCloudwatchConfig, opts ...pulumi.ResourceOption) (*LambdaCloudwatchComponent, error) {
	if args == nil || args.Function == nil {
		return nil, fmt.Errorf("function must be provided")
	}
	comp := &LambdaCloudwatchComponent{}
	err := ctx.RegisterComponentResource("gs:cloudwatch:CloudwatchComponent", name, comp, opts...)
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

	_, err = cloudwatch.NewMetricAlarm(ctx, fmt.Sprintf("%s-no-invocations-alarm", name), &cloudwatch.MetricAlarmArgs{
		AlarmDescription: args.Function.Name.ApplyT(func(fn string) string {
			return fmt.Sprintf("The %s function did not run", fn)
		}).(pulumi.StringOutput),
		ComparisonOperator: pulumi.String("LessThanThreshold"),
		EvaluationPeriods:  pulumi.Int(1),
		Threshold:          pulumi.Float64(1),
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
	}, options...)
	if err != nil {
		return nil, err
	}

	schedulerAssumePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{"sts:AssumeRole"},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					iam.GetPolicyDocumentStatementPrincipal{
						Type:        "Service",
						Identifiers: []string{"scheduler.amazonaws.com"},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	schedulerRole, err := iam.NewRole(ctx, fmt.Sprintf("%s-scheduler-role", name), &iam.RoleArgs{
		Name:             args.Function.Name.ApplyT(func(n string) string { return fmt.Sprintf("%s-schedule-role", n) }).(pulumi.StringOutput),
		AssumeRolePolicy: pulumi.String(schedulerAssumePolicy.Json),
	}, options...)
	if err != nil {
		return nil, err
	}

	schedulerInvokePolicyJSON := args.Function.Arn.ApplyT(func(arn string) (string, error) {
		policyDoc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Actions:   []string{"lambda:Invoke*"},
					Resources: []string{arn},
				},
			},
		}, nil)
		if err != nil {
			return "", err
		}
		return policyDoc.Json, nil
	}).(pulumi.StringOutput)

	schedulerInvokePolicy, err := iam.NewPolicy(ctx, fmt.Sprintf("%s-invoke-policy", name), &iam.PolicyArgs{
		Policy: schedulerInvokePolicyJSON,
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-invoke-policy-attach", name), &iam.RolePolicyAttachmentArgs{
		Role:      schedulerRole.Name,
		PolicyArn: schedulerInvokePolicy.Arn,
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = scheduler.NewSchedule(ctx, fmt.Sprintf("%s-scheduler-schedule", name), &scheduler.ScheduleArgs{
		Name: args.Function.Name.ApplyT(func(n string) string { return fmt.Sprintf("%s-schedule", n) }).(pulumi.StringOutput),
		FlexibleTimeWindow: scheduler.ScheduleFlexibleTimeWindowArgs{
			Mode: pulumi.String("OFF"),
		},
		ScheduleExpression:         pulumi.String("cron(30 2 * * ? *)"),
		ScheduleExpressionTimezone: pulumi.String("America/New_York"),
		State:                      pulumi.String("ENABLED"),
		Target: scheduler.ScheduleTargetArgs{
			Arn:     args.Function.Arn,
			RoleArn: schedulerRole.Arn,
			RetryPolicy: scheduler.ScheduleTargetRetryPolicyArgs{
				MaximumRetryAttempts:     pulumi.Int(args.RetryLimit),
				MaximumEventAgeInSeconds: pulumi.Int(args.Timeout),
			},
		},
	}, options...)
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
