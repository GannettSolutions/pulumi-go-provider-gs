package s3_test

import (
	"testing"

	s3 "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/s3"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/testutil"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestS3BucketComponent(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		key, err := kms.NewKey(ctx, "k", nil)
		if err != nil {
			return err
		}
		_, err = s3.NewS3BucketComponent(ctx, "bucket", &s3.Args{KmsKeyArn: key.Arn})
		return err
	}, pulumi.WithMocks("proj", "stack", &testutil.TestMocks{}))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
