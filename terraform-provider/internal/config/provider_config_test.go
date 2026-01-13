package config

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderConfig_ToClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ProviderConfig
		wantErr bool
	}{
		{
			name: "missing endpoint",
			config: ProviderConfig{
				Endpoint: types.StringNull(),
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: ProviderConfig{
				Endpoint: types.StringValue("http://localhost:8080"),
				Token:    types.StringValue("token"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := tt.config.ToClient(context.Background())

			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.wantErr && client == nil {
				t.Fatalf("expected client")
			}
		})
	}
}
