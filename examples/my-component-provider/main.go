// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package main shows how a simple comoponent provider can be created using existing
// Pulumi programs that contain components.
package main

import (
	"context"
	"fmt"
	"os"

	apigateway "github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/api_gateway"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/cloudwatch"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/ecr"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/lambda"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/ms_entra"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/internal/s3"
	"github.com/GannettSolutions/pulumi-go-provider-gs/examples/my-component-provider/nested"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

func main() {
	provider, err := provider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}

	err = provider.Run(context.Background(), "go-components", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

func provider() (p.Provider, error) {
	return infer.NewProviderBuilder().
		WithNamespace("examples").
		WithComponents(
			infer.ComponentF(NewMyComponent),
			infer.ComponentF(nested.NewNestedRandomComponent),
			infer.ComponentF(lambda.NewLambdaComponent),
			infer.ComponentF(ecr.NewEcrComponent),
			infer.ComponentF(ms_entra.NewEntraAppComponent),
			infer.ComponentF(apigateway.NewApiGatewayLambdaComponent),
			infer.ComponentF(cloudwatch.NewLogGroupComponent),
			infer.ComponentF(cloudwatch.NewLambdaCloudwatchComponent),
			infer.ComponentF(s3.NewS3BucketComponent),
		).
		Build()
}
