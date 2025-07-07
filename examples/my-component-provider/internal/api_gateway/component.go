// Package apigateway contains a component that exposes a simple REST API backed by a Lambda function.
package apigateway

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Args configure NewApiGatewayLambdaComponent.
type Args struct {
	Function *awslambda.Function
	Provider *aws.Provider
}

// ApiGatewayLambdaComponent exposes a simple REST API backed by a Lambda function.
type ApiGatewayLambdaComponent struct {
	pulumi.ResourceState

	URL pulumi.StringOutput
}

func NewApiGatewayLambdaComponent(ctx *pulumi.Context, name string, args *Args, opts ...pulumi.ResourceOption) (*ApiGatewayLambdaComponent, error) {
	if args == nil || args.Function == nil {
		return nil, fmt.Errorf("Function is required")
	}

	comp := &ApiGatewayLambdaComponent{}
	if err := ctx.RegisterComponentResource("oneblood:apigateway:ApiGatewayLambdaComponent", name, comp, opts...); err != nil {
		return nil, err
	}
	options := append(opts, pulumi.Parent(comp))
	if args.Provider != nil {
		options = append(options, pulumi.Provider(args.Provider))
	}

	stack := ctx.Stack()
	project := ctx.Project()
	prefix := fmt.Sprintf("%s-%s-%s", stack, project, name)

	api, err := apigateway.NewRestApi(ctx, fmt.Sprintf("%s-api", prefix), nil, options...)
	if err != nil {
		return nil, err
	}

	resource, err := apigateway.NewResource(ctx, fmt.Sprintf("%s-proxy", prefix), &apigateway.ResourceArgs{
		RestApi:  api.ID(),
		ParentId: api.RootResourceId,
		PathPart: pulumi.String("{proxy+}"),
	}, options...)
	if err != nil {
		return nil, err
	}

	method, err := apigateway.NewMethod(ctx, fmt.Sprintf("%s-method", prefix), &apigateway.MethodArgs{
		RestApi:       api.ID(),
		ResourceId:    resource.ID(),
		HttpMethod:    pulumi.String("ANY"),
		Authorization: pulumi.String("NONE"),
		RequestParameters: pulumi.BoolMap{
			"method.request.path.proxy": pulumi.Bool(true),
		},
	}, options...)
	if err != nil {
		return nil, err
	}

	integration, err := apigateway.NewIntegration(ctx, fmt.Sprintf("%s-integration", prefix), &apigateway.IntegrationArgs{
		RestApi:               api.ID(),
		ResourceId:            resource.ID(),
		HttpMethod:            method.HttpMethod,
		IntegrationHttpMethod: pulumi.String("POST"),
		Type:                  pulumi.String("AWS_PROXY"),
		Uri:                   args.Function.InvokeArn,
	}, options...)
	if err != nil {
		return nil, err
	}

	deployment, err := apigateway.NewDeployment(ctx, fmt.Sprintf("%s-deployment", prefix), &apigateway.DeploymentArgs{
		RestApi: api.ID(),
	}, append(options, pulumi.DependsOn([]pulumi.Resource{integration}))...)
	if err != nil {
		return nil, err
	}

	logGroup, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-logs", prefix), &cloudwatch.LogGroupArgs{
		RetentionInDays: pulumi.Int(30),
	}, options...)
	if err != nil {
		return nil, err
	}

	stage, err := apigateway.NewStage(ctx, fmt.Sprintf("%s-stage", prefix), &apigateway.StageArgs{
		Deployment: deployment.ID(),
		RestApi:    api.ID(),
		StageName:  pulumi.String(stack),
		AccessLogSettings: apigateway.StageAccessLogSettingsArgs{
			DestinationArn: logGroup.Arn,
			Format:         pulumi.String(`{ "requestId":"$context.requestId", "ip":"$context.identity.sourceIp", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "routeKey":"$context.routeKey", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`),
		},
	}, append(options, pulumi.DependsOn([]pulumi.Resource{logGroup}))...)
	if err != nil {
		return nil, err
	}

	_, err = awslambda.NewPermission(ctx, fmt.Sprintf("%s-permission", prefix), &awslambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  args.Function.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.All(api.ID(), stage.StageName).ApplyT(func(vs []interface{}) string {
			return fmt.Sprintf("arn:aws:execute-api:*:*:%v/%v/*", vs[0], vs[1])
		}).(pulumi.StringOutput),
	}, options...)
	if err != nil {
		return nil, err
	}

	invokeOpts := []pulumi.InvokeOption{}
	if args.Provider != nil {
		invokeOpts = append(invokeOpts, pulumi.Provider(args.Provider))
	}
	region := aws.GetRegionOutput(ctx, aws.GetRegionOutputArgs{}, invokeOpts...).Name()
	url := pulumi.All(api.ID(), stage.StageName, region).ApplyT(func(vs []interface{}) string {
		return fmt.Sprintf("https://%v.execute-api.%v.amazonaws.com/%v/", vs[0], vs[2], vs[1])
	}).(pulumi.StringOutput)

	comp.URL = url

	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{"url": url}); err != nil {
		return nil, err
	}

	return comp, nil
}
