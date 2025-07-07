package cloudwatch_test

import (
	"testing"

	cloudwatch "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/cloudwatch"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/testutil"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestCloudwatchComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		fn, err := awslambda.NewFunction(ctx, "fn", &awslambda.FunctionArgs{
			PackageType: pulumi.String("Image"),
			ImageUri:    pulumi.String("example"),
			Role:        pulumi.String("arn:aws:iam::123456789012:role/test"),
		})
		if err != nil {
			return err
		}
		_, err = cloudwatch.NewLambdaCloudwatchComponent(ctx, "cw", &cloudwatch.LambdaCloudwatchConfig{
			Function:                fn,
			LogRetentionDays:        7,
			AlertsEmail:             "test@example.com",
			RetryLimit:              1,
			EnableScheduler:         true,
			EnableAlerts:            true,
			SchedulerCronExpression: "cron(0 12 * * ? *)",
			SchedulerTimezone:       "America/New_York",
		})
		return err
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
