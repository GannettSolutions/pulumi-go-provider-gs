package lambdaapigateway

import (
	"fmt"

	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/apigateway"
	"github.com/OneBloodDataScience/pulumi-oneblood/go/internal/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	command "github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Args configure NewLambdaApiGatewayComponent.
type Args struct {
	lambda.LambdaComponentArgs
	Provider *aws.Provider
}

// LambdaApiGatewayComponent provisions a Lambda and an API Gateway fronting it.
type LambdaApiGatewayComponent struct {
	pulumi.ResourceState

	Lambda  *lambda.LambdaComponent
	Gateway *apigateway.ApiGatewayLambdaComponent
	URL     pulumi.StringOutput
}

func NewLambdaApiGatewayComponent(ctx *pulumi.Context, name string, args *Args, opts ...pulumi.ResourceOption) (*LambdaApiGatewayComponent, error) {
	if args == nil {
		args = &Args{}
	}
	comp := &LambdaApiGatewayComponent{}
	if err := ctx.RegisterComponentResource("oneblood:lambda:LambdaApiGatewayComponent", name, comp, opts...); err != nil {
		return nil, err
	}
	options := append(opts, pulumi.Parent(comp))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	lambdaComp, err := lambda.NewLambdaComponent(ctx, fmt.Sprintf("%s-lambda", name), &args.LambdaComponentArgs, options...)
	if err != nil {
		return nil, err
	}

	api, err := apigateway.NewApiGatewayLambdaComponent(ctx, fmt.Sprintf("%s-api", name), &apigateway.Args{
		Function: lambdaComp.Function,
		Provider: args.Provider,
	}, options...)
	if err != nil {
		return nil, err
	}

	// Update lambda environment with API URL using AWS CLI
	_, err = command.NewCommand(ctx, fmt.Sprintf("%s-setenv", name), &command.CommandArgs{
		Create: pulumi.All(api.URL, lambdaComp.Function.Name).ApplyT(func(vs []interface{}) (string, error) {
			url := vs[0].(string)
			fname := vs[1].(string)
			return fmt.Sprintf("aws lambda update-function-configuration --function-name %s --environment Variables=API_URL=%s", fname, url), nil
		}).(pulumi.StringOutput),
	}, append(options, pulumi.DependsOn([]pulumi.Resource{api}))...)
	if err != nil {
		return nil, err
	}

	comp.Lambda = lambdaComp
	comp.Gateway = api
	comp.URL = api.URL

	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{"url": api.URL}); err != nil {
		return nil, err
	}
	return comp, nil
}
