package testutil

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestMocks implements the pulumi.MockResourceMonitor interface for testing purposes.

type TestMocks struct{}

func (m *TestMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	id := args.Name + "_id"
	outputs := args.Inputs
	switch args.TypeToken {
	case "aws:iam/role:Role":
		outputs["arn"] = resource.NewStringProperty("arn:aws:iam::123456789012:role/" + args.Name)
	case "aws:iam/openIdConnectProvider:OpenIdConnectProvider":
		outputs["arn"] = resource.NewStringProperty("arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com")
	case "aws:iam/rolePolicy:RolePolicy":
		outputs["arn"] = resource.NewStringProperty("arn:aws:iam::123456789012:policy/" + args.Name)
	case "command:local:Command":
		outputs["stdout"] = resource.NewStringProperty("ok")
	case "aws:ecr/repository:Repository":
		outputs["arn"] = resource.NewStringProperty("arn:aws:ecr:us-east-1:123456789012:repository/" + args.Name)
	case "aws:lambda/function:Function":
		outputs["arn"] = resource.NewStringProperty("arn:aws:lambda:us-east-1:123456789012:function:" + args.Name)
	case "aws:lambda/permission:Permission":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:s3/bucketNotification:BucketNotification":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:cloudwatch/logGroup:LogGroup":
		outputs["arn"] = resource.NewStringProperty("arn:aws:logs:us-east-1:123456789012:log-group:" + args.Name)
	case "aws:cloudwatch/metricAlarm:MetricAlarm":
		outputs["arn"] = resource.NewStringProperty("arn:aws:cloudwatch:us-east-1:123456789012:alarm:" + args.Name)
	case "aws:sns/topic:Topic":
		outputs["arn"] = resource.NewStringProperty("arn:aws:sns:us-east-1:123456789012:" + args.Name)
	case "aws:sns/topicSubscription:TopicSubscription":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:s3/bucket:Bucket":
		outputs["bucket"] = resource.NewStringProperty(args.Name)
	case "aws:s3/bucketPublicAccessBlock:BucketPublicAccessBlock":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:apigateway/restApi:RestApi":
		outputs["id"] = resource.NewStringProperty(args.Name)
		outputs["rootResourceId"] = resource.NewStringProperty("root")
	case "aws:apigateway/resource:Resource":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:apigateway/method:Method":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:apigateway/integration:Integration":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:apigateway/deployment:Deployment":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "aws:apigateway/stage:Stage":
		outputs["id"] = resource.NewStringProperty(args.Name)
	case "azuread:index/application:Application":
		outputs["applicationId"] = resource.NewStringProperty("app-id-" + args.Name)
		outputs["objectId"] = resource.NewStringProperty("obj-" + args.Name)
	case "azuread:index/servicePrincipal:ServicePrincipal":
		outputs["id"] = resource.NewStringProperty("sp-" + args.Name)
	case "azuread:index/applicationPassword:ApplicationPassword":
		outputs["value"] = resource.NewStringProperty("secret-value-" + args.Name)
	case "aws:secretsmanager/secret:Secret":
		outputs["arn"] = resource.NewStringProperty("arn:aws:secretsmanager:us-east-1:123456789012:secret:" + args.Name)
	case "aws:secretsmanager/secretVersion:SecretVersion":
		outputs["id"] = resource.NewStringProperty(args.Name)
	}
	return id, outputs, nil
}

func (m *TestMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if args.Token == "aws:index/getCallerIdentity:getCallerIdentity" {
		return resource.PropertyMap{
			"accountId": resource.NewStringProperty("123456789012"),
		}, nil
	}
	return resource.PropertyMap{}, nil
}
