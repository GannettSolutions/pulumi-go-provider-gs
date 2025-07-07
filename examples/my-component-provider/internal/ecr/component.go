// Package ecr contains a component that creates an ECR repository.
package ecr

import (
	"fmt"

	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/core"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// EcrComponentArgs defines the inputs for the EcrComponent.
type EcrComponentArgs struct {
	RepoName pulumi.StringInput `pulumi:"repoName"`
	Provider *aws.Provider      `pulumi:"provider"`
}

// EcrComponent is a component resource that creates an ECR repository.
type EcrComponent struct {
	pulumi.ResourceState

	// The created ECR repository.
	Repository *ecr.Repository `pulumi:"repository"`
}

// NewEcrComponent creates a new EcrComponent.
func NewEcrComponent(ctx *pulumi.Context, name string, args *EcrComponentArgs, opts ...pulumi.ResourceOption) (*EcrComponent, error) {
	if args == nil {
		args = &EcrComponentArgs{}
	}

	component := &EcrComponent{}
	err := ctx.RegisterComponentResource("gs:ecr:EcrComponent", name, component, opts...)
	if err != nil {
		return nil, err
	}

	options := append(opts, pulumi.Parent(component))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	repoNameComplete := pulumi.Sprintf("%s/%s/%s", ctx.Project(), ctx.Stack(), args.RepoName)

	// repoNameComplete := pulumi.Sprintf("%s/%s/%s", pulumi.GetStack(), pulumi.GetProject(), args.RepoName)
	// repoNameComplete := pulumi.All(pulumi.GetStack(), pulumi.GetProject(), args.RepoName).ApplyT(func(args []interface{}) (string, error) {
	// 	stack := args[0].(string)
	// 	project := args[1].(string)
	// 	repoName := args[2].(string)
	// 	return fmt.Sprintf("%s/%s/%s", stack, project, repoName), nil
	// }).(pulumi.StringOutput)
	//

	ecrRepo, err := ecr.NewRepository(ctx, fmt.Sprintf("%s-ecr-repo", name), &ecr.RepositoryArgs{
		Name:        repoNameComplete,
		ForceDelete: pulumi.Bool(true),
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		ImageTagMutability: pulumi.String("MUTABLE"),
		Tags:               core.GenerateBaseTags(ctx),
	}, options...)
	if err != nil {
		return nil, err
	}

	component.Repository = ecrRepo

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"repository": ecrRepo,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
