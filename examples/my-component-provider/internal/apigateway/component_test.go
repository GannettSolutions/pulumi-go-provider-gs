package apigateway_test

import (
	"testing"

	apigw "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/apigateway"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestApiGatewayComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		fn, err := awslambda.NewFunction(ctx, "fn", &awslambda.FunctionArgs{
			PackageType: pulumi.String("Image"),
			ImageUri:    pulumi.String("example"),
			Role:        pulumi.String("arn:aws:iam::123456789012:role/test"),
		})
		if err != nil {
			return err
		}
		_, err = apigw.NewApiGatewayLambdaComponent(ctx, "api", &apigw.Args{Function: fn})
		return err
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
