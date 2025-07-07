// Package core contains shared code for the provider.
package core

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// GenerateBaseTags replicates the Python helper to tag resources with
// common environment information.
func GenerateBaseTags(ctx *pulumi.Context) pulumi.StringMap {
	env := ctx.Stack()
	project := ctx.Project()
	prod := "false"
	if env == "prd" {
		prod = "true"
	}
	return pulumi.StringMap{
		"Production":   pulumi.String(prod),
		"Main_Project": pulumi.String(project),
		"Environment":  pulumi.String(env),
	}
}
