package datasource

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/client"
	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/models"
	sentinelSchema "github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

type PolicyStatusDataSource struct {
	client *client.SentinelClient
}

func NewPolicyStatusDataSource() datasource.DataSource {
	return &PolicyStatusDataSource{}
}

func (d *PolicyStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "sentinel_policy_status"
}

func (d *PolicyStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = sentinelSchema.PolicyStatusSchema()
}

func (d *PolicyStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.SentinelClient)
}

func (d *PolicyStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state models.PolicyStatusModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.GetPolicyStatus(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get policy status", err.Error())
		return
	}

	if status == nil {
		resp.Diagnostics.AddError("policy not found", "sentinel returned no status for policy")
		return
	}

	newState := models.FlattenPolicyStatus(status)
	resp.State.Set(ctx, newState)
}
