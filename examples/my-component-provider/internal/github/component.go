package github

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GithubOidcComponent replicates the GithubOidcComponent.
type GithubOidcComponent struct {
	pulumi.ResourceState

	AccountID  pulumi.StringOutput
	RoleArn    pulumi.StringOutput
	GithubRepo string
	Secrets    pulumi.ArrayOutput
}

type GithubOidcComponentArgs struct {
	GithubRepo         string
	EcrRepos           []*ecr.Repository
	GithubBranch       string
	SaveGithubSecrets  bool
	AdditionalRepoArns []string
	Provider           *aws.Provider
}

func NewGithubOidcComponent(ctx *pulumi.Context, name string, args *GithubOidcComponentArgs, opts ...pulumi.ResourceOption) (*GithubOidcComponent, error) {
	if args == nil {
		args = &GithubOidcComponentArgs{}
	}
	if args.GithubBranch == "" {
		args.GithubBranch = "main"
	}

	comp := &GithubOidcComponent{GithubRepo: args.GithubRepo}
	err := ctx.RegisterComponentResource("oneblood:github:OidcComponent", name, comp, opts...)
	if err != nil {
		return nil, err
	}

	optsInvoke := []pulumi.InvokeOption{}
	if args.Provider != nil {
		optsInvoke = append(optsInvoke, pulumi.Provider(args.Provider))
	}
	identity, err := aws.GetCallerIdentity(ctx, optsInvoke...)
	if err != nil {
		return nil, err
	}
	comp.AccountID = pulumi.String(identity.AccountId).ToStringOutput()

	roleOut, err := CreateGithubOidcRole(ctx, args.EcrRepos, args.GithubRepo, args.Provider, args.GithubBranch, args.AdditionalRepoArns)
	if err != nil {
		return nil, err
	}

	comp.RoleArn = roleOut["AWS_OIDC_ROLE_ARN"].(pulumi.StringOutput)

	if args.SaveGithubSecrets {
		comp.Secrets = pulumi.All(comp.AccountID, comp.RoleArn).ApplyT(func(vals []interface{}) (pulumi.ArrayOutput, error) {
			acc := vals[0].(string)
			arn := vals[1].(string)
			return SetGitHubSecrets(ctx, map[string]string{
				"AWS_ACCOUNT_ID":    acc,
				"AWS_OIDC_ROLE_ARN": arn,
			}, args.GithubRepo, "GithubOidcComponent", ctx.DryRun())
		}).(pulumi.ArrayOutput)
	}

	err = ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"account_id":  comp.AccountID,
		"role_arn":    comp.RoleArn,
		"github_repo": pulumi.String(args.GithubRepo),
		"secrets":     comp.Secrets,
	})
	if err != nil {
		return nil, err
	}
	return comp, nil
}
