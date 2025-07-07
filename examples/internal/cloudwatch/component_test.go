package cloudwatch_test

import (
	"testing"

	cloudwatch "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/cloudwatch"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
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
		_, err = cloudwatch.NewCloudwatchComponent(ctx, "cw", &cloudwatch.Args{
			Function:         fn,
			LogRetentionDays: 7,
			AlertsEmail:      "test@example.com",
			RetryLimit:       1,
		})
		return err
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
