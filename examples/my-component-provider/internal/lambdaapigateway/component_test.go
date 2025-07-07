package lambdaapigateway_test

import (
	"testing"

	lambdapkg "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/lambda"
	apigwcomp "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/lambdaapigateway"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestLambdaApiGatewayComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repo, err := ecr.NewRepository(ctx, "repo", nil)
		if err != nil {
			return err
		}
		_, err = apigwcomp.NewLambdaApiGatewayComponent(ctx, "example", &apigwcomp.Args{
			LambdaComponentArgs: lambdapkg.LambdaComponentArgs{
				LambdaConfig: lambdapkg.LambdaConfig{Timeout: 60, MemorySize: 128},
				EcrRepo:      repo,
			},
		})
		return err
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
