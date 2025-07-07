package github

import (
	"fmt"
	"regexp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const oidcURL = "https://token.actions.githubusercontent.com"

var repoRegex = regexp.MustCompile(`^[\w-]+/[\w.-]+$`)

// CreateGithubOidcRole creates the IAM role and policy used by GitHub Actions.
// This mirrors the functionality of the Python implementation.
func CreateGithubOidcRole(ctx *pulumi.Context,
	ecrRepos []*ecr.Repository,
	githubRepo string,
	provider *aws.Provider,
	githubBranch string,
	additionalRepoArns []string,
) (map[string]pulumi.Output, error) {

	if !repoRegex.MatchString(githubRepo) {
		return nil, fmt.Errorf("Invalid GitHub repo format. Must be 'owner/repo'")
	}
	if len(ecrRepos) == 0 {
		return nil, fmt.Errorf("ecr_repos must be provided")
	}

	for _, r := range ecrRepos {
		if r == nil {
			return nil, fmt.Errorf("all items in ecr_repos must be aws.ecr.Repository instances")
		}
	}

	project, stack := ctx.Project(), ctx.Stack()
	namePrefix := fmt.Sprintf("%s-%s", stack, project)

	opts := []pulumi.ResourceOption{}
	if provider != nil {
		opts = append(opts, pulumi.Provider(provider))
	}
	oidcProvider, err := iam.NewOpenIdConnectProvider(ctx, fmt.Sprintf("%s-github-actions-oidc", namePrefix), &iam.OpenIdConnectProviderArgs{
		Url:             pulumi.String(oidcURL),
		ClientIdLists:   pulumi.StringArray{pulumi.String("sts.amazonaws.com")},
		ThumbprintLists: pulumi.StringArray{pulumi.String("6938fd4d98bab03faadb97b34396831e3780aea1")},
	}, opts...)
	if err != nil {
		return nil, err
	}

	var repoArns []pulumi.StringInput
	for _, r := range ecrRepos {
		repoArns = append(repoArns, r.Arn)
	}
	for _, a := range additionalRepoArns {
		repoArns = append(repoArns, pulumi.String(a))
	}

	oidcSub := fmt.Sprintf("repo:%s:ref:refs/heads/%s", githubRepo, githubBranch)

	role, err := iam.NewRole(ctx, fmt.Sprintf("%s-github-actions-role", namePrefix), &iam.RoleArgs{
		AssumeRolePolicy: oidcProvider.Arn.ApplyT(func(arn string) (string, error) {
			pol := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Federated":"%s"},"Action":"sts:AssumeRoleWithWebIdentity","Condition":{"StringEquals":{"token.actions.githubusercontent.com:aud":"sts.amazonaws.com","token.actions.githubusercontent.com:sub":"%s"}}}]}`,
				arn, oidcSub)
			return pol, nil
		}).(pulumi.StringOutput),
	}, opts...)
	if err != nil {
		return nil, err
	}

	in := make([]interface{}, len(repoArns))
	for i, v := range repoArns {
		in[i] = v
	}

	policy, err := iam.NewRolePolicy(ctx, fmt.Sprintf("%s-github-actions-ecr-policy", namePrefix), &iam.RolePolicyArgs{
		Role: role.ID(),
		Policy: pulumi.All(in...).ApplyT(func(arns []interface{}) (string, error) {
			var s []string
			for _, a := range arns {
				s = append(s, a.(string))
			}
			policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["ecr:GetAuthorizationToken"],"Resource":"*"},{"Effect":"Allow","Action":["ecr:BatchCheckLayerAvailability","ecr:CompleteLayerUpload","ecr:GetDownloadUrlForLayer","ecr:InitiateLayerUpload","ecr:PutImage","ecr:UploadLayerPart","ecr:ListImages","ecr:DescribeImages","ecr:BatchGetImage","ecr:DescribeRepositories"],"Resource":%q}]}`,
				s)
			return policy, nil
		}).(pulumi.StringOutput),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return map[string]pulumi.Output{
		"AWS_OIDC_ROLE_ARN": role.Arn,
		"AWS_OIDC_POLCY":    policy.ToRolePolicyOutput(), // keep typo from Python
	}, nil
}
