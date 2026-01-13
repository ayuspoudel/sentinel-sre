package resource

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/client"
	"github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/models"
	sentinelPolicySchema "github.com/ayuspoudel/sentinel-sre/terraform-provider/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type PolicyResource struct {
	client *client.SentinelClient
}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

/*
Author: @ayuspoudel
Provider has the data for client stored in ProviderData inside req struct.
This needs to be passed to resource as req.ProviderData.(*client.SentinelClient)
so r.client is initialized with the client needed.
At this point our PolicyResource will have all the methods defined for creation
deletion, update and get of policies from sentinel which was the details abstracted
by the client.
*/
func (r *PolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.SentinelClient)
}

/*
Author: @ayuspoudel
All this does is set metadata for resource name.
This is similar to aws_instance or aws_iam_profile and so on.
*/
func (r *PolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "sentinel_policy"
}

/*
Author: @ayuspoudel
At this function we are passing the definitions we have created in
schema/policy_schema.go for sentinel policy resource. This will let
resource know it has this following schema valid, and users should
pass the inputs. It also has optional or defined fields already fulfiled.
*/
func (r *PolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Required: true},
		},
		Blocks: map[string]schema.Block{
			"metadata": sentinelPolicySchema.MetadataSchema(),
			"target":   sentinelPolicySchema.TargetSchema(),
			"signals":  sentinelPolicySchema.SignalsSchema(),
			"policy":   sentinelPolicySchema.PolicySchema(),
		},
	}
}

/*
Author: @ayuspoudel
This function creates the actual resource policy we have defined.
It uses request given by terraform from user input, which we store in our
variable of type PolicyModel which is defined based on terraform types. Then
we see if there are any errors else respond to terraform with the error.
Then we use client.ApplyPolicy(...) where we expand the PolicyModel into
sentinel understandable PolicyModel using our expand function and if apply is
successful we set the state file with the contents of our plan.
*/
func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// We create a plan variable of type PolicyModel to put user defined terraform config
	var plan models.PolicyModel
	// req.Plan.Get gives us the user defined config and puts it in &plan
	diags := req.Plan.Get(ctx, &plan)
	// req.Plan.Get responds with all diagnostics terraform has may it be error, success and so on
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Convert the plan to spec.PolicySpec which is understood by sentinel client
	policySpec := models.ExpandPolicySpec(plan)
	// Use the client to apply the policy to sentinel
	err := r.client.ApplyPolicy(ctx, plan.Name.ValueString(), policySpec)
	// If we get error we append it to diagnostics
	if err != nil {
		resp.Diagnostics.AddError("failed to create sentinel_policy", err.Error())
		return
	}
	// After successful apply this line tells terraform to set the state file with contents from plan
	resp.State.Set(ctx, plan)

}

/*
Author: @ayuspoudel
This is used by terraform to get the current details of the policy via the GET endpoint provided
by sentinel and abstracted by our client. So it calles client.GetPolicy(policyName). It responds
in *spec.PolicySpec struct so we use policy_flatten() function to convert it to terraform understandable
policyModel and if anything mismatches we set the current state to be of the new state.
*/
func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// We initialize a policy model to put what terraform has from request (resource.ReadRequest)
	var state models.PolicyModel
	// Store the current policy state in state var
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Use client to get the current policy from sentinel
	policySpec, err := r.client.GetPolicy(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get current policy", err.Error())
		return
	}
	if policySpec == nil {
		// Policy not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}
	// Convert the spec.PolicySpec to terraform understandable PolicyModel
	updatedState := models.FlattenPolicySpec(policySpec)
	// Set the state to updated state
	resp.State.Set(ctx, updatedState)

}

/*
Author: @ayuspoudel
Sentinel API is idempotent so internally both update and create function contents remain
the same.
*/
func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.PolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policySpec := models.ExpandPolicySpec(plan)
	err := r.client.ApplyPolicy(ctx, plan.Name.ValueString(), policySpec)
	if err != nil {
		resp.Diagnostics.AddError("failed to update sentinel_policy", err.Error())
		return
	}
	resp.State.Set(ctx, plan)

}

/*
Author: @ayuspoudel
This function is used by terraform to delete any resources. For this we simply
get the current state from req resource.DeleteRequest's State.Get() method
and use client.DeletePolicy(...) method to delete the policy from sentinel.
*/
func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.PolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeletePolicy(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete sentinel_policy", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)

}
