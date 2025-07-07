package github_test

import (
	"strings"
	"testing"

	github "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/github"
	testutil "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/testutil"
	github_provider "github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestComponentInitialization(t *testing.T) {

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repo, err := ecr.NewRepository(ctx, "repo", nil)
		if err != nil {
			return err
		}
		githubProvider, err := github_provider.NewProvider(ctx, "github-provider", &github_provider.ProviderArgs{
			Owner: pulumi.String("test-org"),
		})
		if err != nil {
			return err
		}

		comp, err := github.NewGithubOidcComponent(ctx, "test", &github.GithubOidcComponentArgs{
			GithubRepo:        "test-org/test-repo",
			EcrRepos:          []*ecr.Repository{repo},
			SaveGithubSecrets: true,
			SaveGithubVariables: true,
			GithubProvider:    githubProvider,
		})
		if err != nil {
			return err
		}
		comp.GithubRepo.ApplyT(func(r string) error {
			if r != "test-org/test-repo" {
				t.Errorf("repo mismatch: %s", r)
			}
			return nil
		})
		comp.RoleArn.ApplyT(func(a string) error {
			if !strings.HasPrefix(a, "arn:aws:iam::123456789012:role/") {
				t.Errorf("bad arn: %s", a)
			}
			return nil
		})
		comp.Secrets.ApplyT(func(arr []interface{}) error {
			if len(arr) != 2 {
				t.Errorf("expected 2 secrets, got %d", len(arr))
			}
			return nil
		})
		comp.Variables.ApplyT(func(arr []interface{}) error {
			if len(arr) != 2 {
				t.Errorf("expected 2 variables, got %d", len(arr))
			}
			return nil
		})
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
