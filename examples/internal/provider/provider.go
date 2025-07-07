package provider

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Provider is a placeholder Pulumi component resource implemented in Go.
type Provider struct {
	pulumi.ResourceState
}

// NewProvider registers a minimal component resource. This implementation does
// not replicate the functionality of the Python provider and is meant only as a
// stub example.
func NewProvider(ctx *pulumi.Context, name string, opts ...pulumi.ResourceOption) (*Provider, error) {
	p := &Provider{}
	err := ctx.RegisterComponentResource("oneblood:index:Provider", name, p, opts...)
	if err != nil {
		return nil, err
	}
	if err := ctx.RegisterResourceOutputs(p, pulumi.Map{}); err != nil {
		return nil, err
	}
	return p, nil
}
