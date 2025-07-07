package github_test

import (
	"strings"
	"testing"

	github "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/github"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestInvalidRepoFormat(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := github.CreateGithubOidcRole(ctx, []*ecr.Repository{}, "invalid", nil, "main", nil)
		if err == nil {
			t.Errorf("expected error")
		}
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestInvalidEcrRepos(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := github.CreateGithubOidcRole(ctx, []*ecr.Repository{nil}, "test/repo", nil, "main", nil)
		if err == nil {
			t.Errorf("expected error")
		}
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestValidOidcSetup(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repo, err := ecr.NewRepository(ctx, "repo", nil)
		if err != nil {
			return err
		}
		out, err := github.CreateGithubOidcRole(ctx, []*ecr.Repository{repo}, "test-org/test-repo", nil, "dev", nil)
		if err != nil {
			return err
		}
		arn := out["AWS_OIDC_ROLE_ARN"].(pulumi.StringOutput)
		arn.ApplyT(func(a string) error {
			if !strings.HasPrefix(a, "arn:aws:iam::123456789012:role/") {
				t.Errorf("unexpected arn %s", a)
			}
			return nil
		})
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
