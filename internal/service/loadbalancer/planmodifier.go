package loadbalancer

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getNetworksPlanModifier() planmodifier.List {
	return &networksPlanModifier{}
}

type networksPlanModifier struct{}

func (d *networksPlanModifier) Description(_ context.Context) string {
	return "Ensures network ID is set only if the network type is private."
}

func (d *networksPlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *networksPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var networkModels []loadbalancerNetworkModel
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &networkModels, false)...)

	diags := validateNetworks(networkModels)
	resp.Diagnostics.Append(diags...)
}

func validateNetworks(networkModels []loadbalancerNetworkModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i, n := range networkModels {
		switch upcloud.LoadBalancerNetworkType(n.Type.ValueString()) {
		case upcloud.LoadBalancerNetworkTypePrivate:
			if n.Network.IsNull() {
				diags.AddError("load balancer's private network ID is required", fmt.Sprintf("#%d", i))
			}
		case upcloud.LoadBalancerNetworkTypePublic:
			if !n.Network.IsNull() {
				diags.AddError("setting load balancer's public network ID is not supported", fmt.Sprintf("#%d", i))
			}
		}
	}

	return diags
}

const networkRequiresReplaceIfDescription = "clearing the network field is only allowed when equivalent `networks` block is defined"

func networkRequiresReplaceIfFunc(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	prev := req.StateValue.ValueString()
	next := req.PlanValue.ValueString()

	if prev == "" || next != "" {
		resp.RequiresReplace = true
		return
	}

	var plan loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Networks.IsNull() {
		resp.RequiresReplace = true
		return
	}

	var networkModels []loadbalancerNetworkModel
	resp.Diagnostics.Append(plan.Networks.ElementsAs(ctx, &networkModels, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Equivalent `networks` block contains two entries. First for public network and second for private network where UUID should match network value.
	if len(networkModels) != 2 || networkModels[1].ID.ValueString() != prev {
		resp.RequiresReplace = true
	}

	resp.RequiresReplace = false
}
