package schema

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func PolicyStatusSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},

			"cluster_exists":    schema.BoolAttribute{Computed: true},
			"cluster_reachable": schema.BoolAttribute{Computed: true},
			"namespace_exists":  schema.BoolAttribute{Computed: true},

			"agent_installed": schema.BoolAttribute{Computed: true},
			"agent_healthy":   schema.BoolAttribute{Computed: true},

			"prometheus_reachable": schema.BoolAttribute{Computed: true},
			"queries_valid":        schema.BoolAttribute{Computed: true},

			"last_validated_at": schema.StringAttribute{Computed: true},
			"last_error":        schema.StringAttribute{Computed: true},
		},
	}
}
