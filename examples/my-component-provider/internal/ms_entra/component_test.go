package ms_entra_test

import (
	"testing"

	entra "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/ms_entra"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestEntraAppComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := entra.NewEntraAppComponent(ctx, "test", &entra.Args{
			Domain:       "example.com",
			TenantID:     "tid",
			TenantDomain: "tenant.example.com",
			Provider:     &aws.Provider{},
		})
		return err
	}, pulumi.WithMocks("proj", "dev", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestEntraDisallowLocalhost(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := entra.NewEntraAppComponent(ctx, "test", &entra.Args{
			Domain:       "localhost",
			TenantID:     "tid",
			TenantDomain: "tenant",
			Provider:     &aws.Provider{},
		})
		if err == nil {
			t.Errorf("expected error for localhost domain")
		}
		return nil
	}, pulumi.WithMocks("proj", "prd", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
