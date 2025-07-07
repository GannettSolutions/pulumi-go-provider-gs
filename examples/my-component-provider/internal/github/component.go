package github

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	github_provider "github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GithubOidcComponent replicates the GithubOidcComponent.
type GithubOidcComponent struct {
	pulumi.ResourceState

	AccountID  pulumi.StringOutput
	RoleArn    pulumi.StringOutput
	GithubRepo string
	Secrets    pulumi.ArrayOutput
	Variables  pulumi.ArrayOutput
}

type GithubOidcComponentArgs struct {
	GithubRepo          string
	EcrRepos            []*ecr.Repository
	GithubBranch        string
	SaveGithubSecrets   bool
	SaveGithubVariables bool
	AdditionalRepoArns  []string
	AwsProvider         *aws.Provider
	GithubProvider      *github_provider.Provider
}

func NewGithubOidcComponent(ctx *pulumi.Context, name string, args *GithubOidcComponentArgs, opts ...pulumi.ResourceOption) (*GithubOidcComponent, error) {
	if args == nil {
		args = &GithubOidcComponentArgs{}
	}
	if args.GithubBranch == "" {
		args.GithubBranch = "main"
	}

	comp := &GithubOidcComponent{GithubRepo: args.GithubRepo}
	err := ctx.RegisterComponentResource("gs:github:OidcComponent", name, comp, opts...)
	if err != nil {
		return nil, err
	}

	optsInvoke := []pulumi.InvokeOption{}
	if args.AwsProvider != nil {
		optsInvoke = append(optsInvoke, pulumi.Provider(args.AwsProvider))
	}
	identity, err := aws.GetCallerIdentity(ctx, &aws.GetCallerIdentityArgs{}, optsInvoke...)
	if err != nil {
		return nil, err
	}
	comp.AccountID = pulumi.String(identity.AccountId).ToStringOutput()

	roleOut, err := CreateGithubOidcRole(ctx, args.EcrRepos, args.GithubRepo, args.AwsProvider, args.GithubBranch, args.AdditionalRepoArns)
	if err != nil {
		return nil, err
	}

	comp.RoleArn = roleOut["AWS_OIDC_ROLE_ARN"].(pulumi.StringOutput)

	githubOpts := []pulumi.ResourceOption{}
	if args.GithubProvider != nil {
		githubOpts = append(githubOpts, pulumi.Provider(args.GithubProvider))
	}

	if args.SaveGithubSecrets {
		comp.Secrets = pulumi.All(comp.AccountID, comp.RoleArn).ApplyT(func(vals []interface{}) (pulumi.ArrayOutput, error) {
			acc := vals[0].(string)
			arn := vals[1].(string)
			return SetGitHubSecrets(ctx, pulumi.StringMap{
				"AWS_ACCOUNT_ID":    pulumi.String(acc),
				"AWS_OIDC_ROLE_ARN": pulumi.String(arn),
			}, args.GithubRepo, "GithubOidcComponent", githubOpts...)
		}).(pulumi.ArrayOutput)
	}

	if args.SaveGithubVariables {
		comp.Variables = pulumi.All(comp.AccountID, comp.RoleArn).ApplyT(func(vals []interface{}) (pulumi.ArrayOutput, error) {
			acc := vals[0].(string)
			arn := vals[1].(string)
			return SetGitHubVariables(ctx, pulumi.StringMap{
				"AWS_ACCOUNT_ID":    pulumi.String(acc),
				"AWS_OIDC_ROLE_ARN": pulumi.String(arn),
			}, args.GithubRepo, "GithubOidcComponent", githubOpts...)
		}).(pulumi.ArrayOutput)
	}

	err = ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"account_id":  comp.AccountID,
		"role_arn":    comp.RoleArn,
		"github_repo": pulumi.String(comp.GithubRepo),
		"secrets":     comp.Secrets,
		"variables":   comp.Variables,
	})
	if err != nil {
		return nil, err
	}
	return comp, nil
}
