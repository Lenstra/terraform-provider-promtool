// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	successExitCode = 0
	failureExitCode = 1

	lintOptionAll            = "all"
	lintOptionDuplicateRules = "duplicate-rules"
	lintOptionNone           = "none"
	checkHealth              = "/-/healthy"
	checkReadiness           = "/-/ready"
)

// Ensure PromtoolProvider satisfies various provider interfaces.
var _ provider.Provider = &PromtoolProvider{}
var _ provider.ProviderWithFunctions = &PromtoolProvider{}

// PromtoolProvider defines the provider implementation.
type PromtoolProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PromtoolProviderModel describes the provider data model.
type PromtoolProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *PromtoolProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "promtool"
	resp.Version = p.version
}

func (p *PromtoolProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			//TODO : add linter configuration
		},
	}
}

func (p *PromtoolProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PromtoolProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PromtoolProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *PromtoolProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *PromtoolProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewCheckRulesFunction,
		NewCheckConfigFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PromtoolProvider{
			version: version,
		}
	}
}
