package config

import (
	"context"
	"fmt"

	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

/*
Author: @ayuspoudel

ProviderConfig defines configuration accepted by the Sentinel provider.
This struct is the single source of truth for provider-level settings.
*/
type ProviderConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

/*
Author: @ayuspoudel
ToClient validates provider config and constructs a SentinelClient.
This is the only place where:
- endpoint is checked
- token is interpreted
- client creation happens
*/
func (p *ProviderConfig) ToClient(ctx context.Context) (*client.SentinelClient, error) {
	// Endpoint is required at schema level, but we still guard here
	if p.Endpoint.IsNull() || p.Endpoint.ValueString() == "" {
		return nil, fmt.Errorf("sentinel provider: endpoint must be set")
	}
	var token string
	if !p.Token.IsNull() {
		token = p.Token.ValueString()
	}
	return client.NewSentinelClient(p.Endpoint.ValueString(), token), nil
}
