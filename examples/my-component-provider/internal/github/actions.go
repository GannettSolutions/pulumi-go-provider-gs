// Package github contains a component that sets secrets in a GitHub repository using the pulumi-github provider.
// It sets up aws oidc roles and policies for the repository.
package github

import (
	"fmt"

	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SetGitHubSecrets sets secrets in a GitHub repository using the pulumi-github provider.
func SetGitHubSecrets(ctx *pulumi.Context, secrets pulumi.StringMap, repo, componentName string, opts ...pulumi.ResourceOption) (pulumi.ArrayOutput, error) {
	var secretsArr []pulumi.Resource
	for k, v := range secrets {
		secret, err := github.NewActionsSecret(ctx, fmt.Sprintf("%s-%s", componentName, k), &github.ActionsSecretArgs{
			Repository:     pulumi.String(repo),
			SecretName:     pulumi.String(k),
			PlaintextValue: v,
		}, opts...)
		if err != nil {
			return pulumi.ArrayOutput{}, err
		}
		secretsArr = append(secretsArr, secret)
	}

	return pulumi.ToOutput(secretsArr).(pulumi.ArrayOutput), nil
}

// SetGitHubVariables sets variables in a GitHub repository using the pulumi-github provider.
func SetGitHubVariables(ctx *pulumi.Context, variables pulumi.StringMap, repo, componentName string, opts ...pulumi.ResourceOption) (pulumi.ArrayOutput, error) {
	var variablesArr []pulumi.Resource
	for k, v := range variables {
		variable, err := github.NewActionsVariable(ctx, fmt.Sprintf("%s-%s", componentName, k), &github.ActionsVariableArgs{
			Repository:   pulumi.String(repo),
			VariableName: pulumi.String(k),
			Value:        v,
		}, opts...)
		if err != nil {
			return pulumi.ArrayOutput{}, err
		}
		variablesArr = append(variablesArr, variable)
	}

	return pulumi.ToOutput(variablesArr).(pulumi.ArrayOutput), nil
}
