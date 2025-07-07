package provider_test

import (
	"testing"

	pbprovider "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type mockMonitor struct{}

func (m *mockMonitor) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

func (m *mockMonitor) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "-id", resource.PropertyMap{}, nil
}

func TestProviderInstantiation(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := pbprovider.NewProvider(ctx, "test")
		return err
	}, pulumi.WithMocks("proj", "stack", &mockMonitor{}))
	if err != nil {
		t.Fatalf("provider run failed: %v", err)
	}
}
