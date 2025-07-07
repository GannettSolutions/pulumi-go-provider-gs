package ecr_test

import (
	"testing"

	ecr "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type mocks int

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestEcrComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := ecr.NewEcrComponent(ctx, "test-ecr", &ecr.EcrComponentArgs{
			RepoName: pulumi.String("my-repo"),
		})
		assert.NoError(t, err)

		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))
	assert.NoError(t, err)
}
