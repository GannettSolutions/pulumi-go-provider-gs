package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	local "github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ExecCommand is used to invoke external commands. It can be overridden in
// tests.
var ExecCommand = exec.CommandContext

// SetGitHubSecrets creates or updates secrets in a repository using the gh CLI.
// Secrets are managed as pulumi-command resources to mirror the Python version.
func SetGitHubSecrets(ctx *pulumi.Context, secrets map[string]string, repo string, namespace string, dryRun bool) (pulumi.ArrayOutput, error) {
	if secrets == nil {
		return pulumi.ArrayOutput{}, fmt.Errorf("secret_dict must be a dict")
	}
	if !repoRegex.MatchString(repo) {
		return pulumi.ArrayOutput{}, fmt.Errorf("Invalid GitHub repo format. Must be 'owner/repo'")
	}
	dryRun = dryRun || ctx.DryRun()

	if _, err := ExecCommand(context.Background(), "gh", "auth", "status").Output(); err != nil {
		return pulumi.ArrayOutput{}, fmt.Errorf("gh CLI not authenticated. Install GitHub CLI and run 'gh auth login': %w", err)
	}

	existing := map[string]bool{}
	out, err := ExecCommand(context.Background(), "gh", "secret", "list", "--repo", repo, "--json", "name").Output()
	if err == nil {
		var arr []map[string]string
		if jsonErr := json.Unmarshal(out, &arr); jsonErr == nil {
			for _, e := range arr {
				existing[e["name"]] = true
			}
		}
	}

	statePrefix := "gh-secret-"
	if namespace != "" {
		statePrefix = namespace + "-gh-secret-"
	}

	managedInState := map[string]bool{}
	for name := range existing {
		if strings.HasPrefix(name, statePrefix) {
			part := strings.TrimPrefix(name, statePrefix)
			key := strings.SplitN(part, "-", 2)[0]
			managedInState[key] = true
		}
	}

	for k := range secrets {
		if existing[k] && !managedInState[k] {
			return pulumi.ArrayOutput{}, fmt.Errorf("Secret '%s' already exists in GitHub but was not created by this namespace", k)
		}
	}

	var resources []pulumi.Output

	for key, value := range secrets {
		secret := pulumi.ToSecret(pulumi.String(value)).(pulumi.StringOutput)
		cmd := secret.ApplyT(func(val string) (*local.Command, error) {
			h := sha256.Sum256([]byte(key + val))
			frag := hex.EncodeToString(h[:])[:8]
			prefix := ""
			if namespace != "" {
				prefix = namespace + "-"
			}
			name := fmt.Sprintf("%sgh-secret-%s-%s", prefix, strings.ToLower(key), frag)
			if dryRun {
				return local.NewCommand(ctx, name, &local.CommandArgs{
					Create: pulumi.String("echo dry-run: skipped create"),
					Delete: pulumi.String("echo dry-run: skipped delete"),
				})
			}
			escaped := strings.ReplaceAll(val, "'", "'\"'\"'")
			return local.NewCommand(ctx, name, &local.CommandArgs{
				Create: pulumi.String(fmt.Sprintf("echo '%s' | gh secret set %s --repo %s", escaped, key, repo)),
				Delete: pulumi.String(fmt.Sprintf("gh secret delete %s --repo %s --yes", key, repo)),
			})
		}).(pulumi.Output)
		resources = append(resources, cmd)
	}

	desired := map[string]bool{}
	for k := range secrets {
		desired[k] = true
	}
	for key := range managedInState {
		if !desired[key] {
			prefix := ""
			if namespace != "" {
				prefix = namespace + "-"
			}
			name := fmt.Sprintf("%sgh-secret-%s-deleted", prefix, strings.ToLower(key))
			deleteCmd := fmt.Sprintf("gh secret delete %s --repo %s --yes", key, repo)
			if dryRun {
				deleteCmd = "echo dry-run: skipped delete"
			}
			c, _ := local.NewCommand(ctx, name, &local.CommandArgs{
				Create: pulumi.String("echo noop"),
				Delete: pulumi.String(deleteCmd),
			})
			resources = append(resources, c.ToCommandOutput())
		}
	}

	ins := make([]interface{}, len(resources))
	for i, v := range resources {
		ins[i] = v
	}
	return pulumi.All(ins...).ApplyT(func(vs []interface{}) []interface{} { return vs }).(pulumi.ArrayOutput), nil
}
