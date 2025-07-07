// Package entra contains a component that creates an Azure AD application and uploads
// the credentials to AWS Secrets Manager.
package ms_entra

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/secretsmanager"
	ad "github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Args defines the inputs for creating an Entra application.
type Args struct {
	Domain       string
	TenantID     string
	TenantDomain string
	Provider     *aws.Provider
}

// EntraAppComponent provisions an Azure AD application and uploads
// the credentials to AWS Secrets Manager.
type EntraAppComponent struct {
	pulumi.ResourceState

	SecretArn pulumi.StringOutput
	AppID     pulumi.StringOutput
}

func NewEntraAppComponent(ctx *pulumi.Context, name string, args *Args, opts ...pulumi.ResourceOption) (*EntraAppComponent, error) {
	if args == nil {
		return nil, fmt.Errorf("args are required")
	}

	stack := ctx.Stack()
	if (stack == "prd" || stack == "stg") && strings.Contains(args.Domain, "localhost") {
		return nil, fmt.Errorf("localhost is not allowed for domain in prd or stg")
	}

	comp := &EntraAppComponent{}
	if err := ctx.RegisterComponentResource("gs:entra:EntraAppComponent", name, comp, opts...); err != nil {
		return nil, err
	}

	options := append(opts, pulumi.Parent(comp))

	app, err := ad.NewApplication(ctx, fmt.Sprintf("%s-app", name), &ad.ApplicationArgs{
		DisplayName: pulumi.String(fmt.Sprintf("oneblood_%s_%s", name, stack)),
		IdentifierUris: pulumi.StringArray{
			pulumi.String(fmt.Sprintf("oneblood-%s-id-%s", name, stack)),
		},
		SignInAudience:              pulumi.String("AzureADMyOrg"),
		FallbackPublicClientEnabled: pulumi.Bool(false),
		Web: ad.ApplicationWebArgs{
			HomepageUrl: pulumi.String(fmt.Sprintf("https://%s", args.Domain)),
			LogoutUrl:   pulumi.String(fmt.Sprintf("https://%s/logout", args.Domain)),
			RedirectUris: pulumi.ToStringArray(func() []string {
				uris := []string{
					fmt.Sprintf("https://%s/auth/callback", args.Domain),
					fmt.Sprintf("https://%s/auth/getAToken", args.Domain),
				}
				if stack != "prd" && stack != "stg" {
					uris = append([]string{
						"http://localhost:5000/auth/getAToken",
						"http://localhost:5000/auth/callback",
					}, uris...)
				}
				return uris
			}()),
			ImplicitGrant: ad.ApplicationWebImplicitGrantArgs{
				AccessTokenIssuanceEnabled: pulumi.Bool(false),
				IdTokenIssuanceEnabled:     pulumi.Bool(false),
			},
		},
		GroupMembershipClaims: pulumi.StringArray{pulumi.String("SecurityGroup"), pulumi.String("DirectoryRole"), pulumi.String("All")},
		FeatureTags: ad.ApplicationFeatureTagArray{
			ad.ApplicationFeatureTagArgs{Enterprise: pulumi.Bool(true), Hide: pulumi.Bool(false)},
		},
		RequiredResourceAccesses: ad.ApplicationRequiredResourceAccessArray{
			ad.ApplicationRequiredResourceAccessArgs{
				ResourceAppId: pulumi.String("00000003-0000-0000-c000-000000000000"),
				ResourceAccesses: ad.ApplicationRequiredResourceAccessResourceAccessArray{
					ad.ApplicationRequiredResourceAccessResourceAccessArgs{Id: pulumi.String("b340eb25-3456-403f-be2f-af7a0d370277"), Type: pulumi.String("Scope")},
					ad.ApplicationRequiredResourceAccessResourceAccessArgs{Id: pulumi.String("e1fe6dd8-ba31-4d61-89e7-88639da4683d"), Type: pulumi.String("Scope")},
					ad.ApplicationRequiredResourceAccessResourceAccessArgs{Id: pulumi.String("bc024368-1153-4739-b217-4326f2e966d0"), Type: pulumi.String("Scope")},
				},
			},
		},
	}, options...)
	if err != nil {
		return nil, err
	}

	_, err = ad.NewServicePrincipal(ctx, fmt.Sprintf("%s-sp", name), &ad.ServicePrincipalArgs{
		ClientId:                  pulumi.ToOutput(app.ApplicationId).(pulumi.StringOutput),
		Tags:                      pulumi.StringArray{pulumi.String(fmt.Sprintf("stack:%s", stack)), pulumi.String("enterprise:true")},
		AppRoleAssignmentRequired: pulumi.Bool(true),
	}, options...)
	if err != nil {
		return nil, err
	}

	secret, err := ad.NewApplicationPassword(ctx, fmt.Sprintf("%s-secret", name), &ad.ApplicationPasswordArgs{
		ApplicationObjectId: app.ObjectId,
		DisplayName:         pulumi.String(fmt.Sprintf("%s_client_secret_%s", name, stack)),
		EndDate:             pulumi.String("2025-08-10T15:48:31Z"),
	}, options...)
	if err != nil {
		return nil, err
	}

	secretName := fmt.Sprintf("%s-entra-credentials-%s", name, stack)
	smSecret, err := secretsmanager.NewSecret(ctx, fmt.Sprintf("%s-sm", name), &secretsmanager.SecretArgs{
		Name:        pulumi.String(secretName),
		Description: pulumi.String(fmt.Sprintf("Microsoft Entra ID credentials for %s %s", name, stack)),
	}, append(options, pulumi.Provider(args.Provider))...)
	if err != nil {
		return nil, err
	}

	metaURL := pulumi.Sprintf("https://login.microsoftonline.com/%s/federationmetadata/2007-06/federationmetadata.xml?appid=%s", args.TenantID, app.ApplicationId)

	secretString := pulumi.All(pulumi.String(args.Domain), pulumi.String(args.TenantDomain), pulumi.String(args.TenantID), app.ApplicationId, secret.Value, app.ObjectId, metaURL).ApplyT(
		func(vals []interface{}) (string, error) {
			m := map[string]string{
				"DOMAIN":                     vals[0].(string),
				"TENANT_DOMAIN":              vals[1].(string),
				"TENANT_ID":                  vals[2].(string),
				"CLIENT_ID":                  vals[3].(string),
				"CLIENT_SECRET":              vals[4].(string),
				"OBJECT_ID":                  vals[5].(string),
				"APP_FEDERATED_METADATA_URL": vals[6].(string),
			}
			b, err := json.Marshal(m)
			return string(b), err
		}).(pulumi.StringOutput)

	_, err = secretsmanager.NewSecretVersion(ctx, fmt.Sprintf("%s-sm-version", name), &secretsmanager.SecretVersionArgs{
		SecretId:      smSecret.ID(),
		SecretString:  secretString.ToStringPtrOutput(),
		VersionStages: pulumi.StringArray{pulumi.String("AWSCURRENT")},
	}, append(options, pulumi.Provider(args.Provider))...)
	if err != nil {
		return nil, err
	}

	comp.SecretArn = smSecret.Arn
	comp.AppID = app.ApplicationId

	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"secretArn": smSecret.Arn,
		"appId":     app.ApplicationId,
	}); err != nil {
		return nil, err
	}
	return comp, nil
}
