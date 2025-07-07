package github_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	github "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/github"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestComponentInitialization(t *testing.T) {
	github.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { github.ExecCommand = exec.CommandContext }()

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		repo, err := ecr.NewRepository(ctx, "repo", nil)
		if err != nil {
			return err
		}
		comp, err := github.NewGithubOidcComponent(ctx, "test", &github.GithubOidcComponentArgs{
			GithubRepo:        "test-org/test-repo",
			EcrRepos:          []*ecr.Repository{repo},
			SaveGithubSecrets: true,
		})
		if err != nil {
			return err
		}
		if comp.GithubRepo != "test-org/test-repo" {
			t.Errorf("repo mismatch")
		}
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
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
