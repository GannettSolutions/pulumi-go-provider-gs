package github_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	github "github.com/OneBloodDataScience/pulumi-oneblood/go/internal/github"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/testutil"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func mockGHList(existing []string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", string(must(json.Marshal(existing))))
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}

func TestSecretCreatedDryRun(t *testing.T) {
	github.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) >= 2 && args[0] == "auth" && args[1] == "status" {
			return exec.CommandContext(ctx, "echo", "")
		}
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { github.ExecCommand = exec.CommandContext }()

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		res, err := github.SetGitHubSecrets(ctx, map[string]string{"FOO": "bar"}, "org/repo", "Test", true)
		if err != nil {
			return err
		}
		res.ApplyT(func(v []interface{}) error {
			if len(v) != 1 {
				t.Errorf("expected 1 command")
			}
			return nil
		})
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestSecretAuthFailure(t *testing.T) {
	github.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) >= 2 && args[0] == "auth" && args[1] == "status" {
			return exec.CommandContext(ctx, "false")
		}
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { github.ExecCommand = exec.CommandContext }()

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := github.SetGitHubSecrets(ctx, map[string]string{"FOO": "bar"}, "org/repo", "Test", true)
		if err == nil {
			t.Errorf("expected auth error")
		}
		return nil
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
