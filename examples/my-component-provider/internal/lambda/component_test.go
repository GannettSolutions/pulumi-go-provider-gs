package lambda_test

import (
	"fmt"
	"testing"

	lambdapkg "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/lambda"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestLambdaComponentBasic(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repo, err := ecr.NewRepository(ctx, "repo", nil)
		if err != nil {
			return err
		}
		comp, err := lambdapkg.NewLambdaComponent(ctx, "test", &lambdapkg.LambdaComponentArgs{
			LambdaConfig: lambdapkg.LambdaConfig{Timeout: 60, MemorySize: 128},
			EcrRepo:      repo,
		})
		if err != nil {
			return err
		}
		if comp.Function == nil {
			return fmt.Errorf("function not created")
		}
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
