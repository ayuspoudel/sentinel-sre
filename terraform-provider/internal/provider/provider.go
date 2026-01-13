package provider

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/config"
	sentinelResource "github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

/*
Author: @ayuspoudel
Doc: https://developer.hashicorp.com/terraform/plugin/framework/providers
This is a implementation of provider interface. Provider interface needs 5 methods explicitly as
said in the documentation such that our struct satisfied the provider interface.
We have defined a custom struct so we can add more fields so our methods can implicitly access all
the fields we may have in the future without breaking terraform provider.
*/
type SentinelProvider struct {
	version string
}

var _ provider.Provider = &SentinelProvider{}

func New() provider.Provider {
	return &SentinelProvider{version: "0.1.0"}
}

func (p *SentinelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sentinel"
	resp.Version = p.version
}
func (p *SentinelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{Required: true},
			"token":    schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

/*
Author: @ayuspoudel
This function is responsible for initializing shared dependencies (such as the
Sentinel API client) using provider configuration and making them available to
all resources and data sources via ProviderData.
*/
func (p *SentinelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg config.ProviderConfig
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	sentinelClient, err := cfg.ToClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to configure sentinel provider", err.Error())
		return
	}
	resp.ResourceData = sentinelClient
	resp.DataSourceData = sentinelClient

}

func (p *SentinelProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *SentinelProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		sentinelResource.NewPolicyResource,
	}
}
