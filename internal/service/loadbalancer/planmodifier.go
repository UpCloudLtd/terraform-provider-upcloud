package loadbalancer

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

func networkReplacedWithNetworks(ctx context.Context, state, plan loadBalancerModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Ensure that `networks` value was null
	if !state.Networks.IsNull() {
		return false, diags
	}

	// Ensure that `network` value is cleared
	prevNetwork := state.Network.ValueString()
	nextNetwork := plan.Network.ValueString()

	if prevNetwork == "" || nextNetwork != "" {
		return false, diags
	}

	var networkModels []loadbalancerNetworkModel
	diags.Append(plan.Networks.ElementsAs(ctx, &networkModels, false)...)
	if diags.HasError() {
		return false, diags
	}

	// Equivalent `networks` block contains two entries. First for public network and second for private network where UUID should match network value.
	if len(networkModels) != 2 {
		return false, diags
	}

	if networkModels[0].Type.ValueString() != string(upcloud.LoadBalancerNetworkTypePublic) {
		return false, diags
	}

	if networkModels[1].Type.ValueString() != string(upcloud.LoadBalancerNetworkTypePrivate) {
		return false, diags
	}

	if networkModels[1].Network.ValueString() != prevNetwork {
		return false, diags
	}

	return true, diags
}

const networkRequiresReplaceIfDescription = "clearing the `network` field is only allowed when equivalent `networks` block is defined"

func networkRequiresReplaceIfFunc(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	replaced, diags := networkReplacedWithNetworks(ctx, state, plan)
	resp.RequiresReplace = !replaced
	resp.Diagnostics.Append(diags...)
}

const networksRequiresReplaceIfDescription = "modifying `networks` field is only allowed when it replaces equivalent `network` value"

func networksRequiresReplaceIfFunc(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state loadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	replaced, diags := networkReplacedWithNetworks(ctx, state, plan)
	resp.Diagnostics.Append(diags...)

	if replaced {
		resp.RequiresReplace = false
		return
	}

	stateLen := len(state.Networks.Elements())
	planLen := len(plan.Networks.Elements())

	if stateLen == planLen {
		// No change in length, nested fields have their own plan modifiers to require replace if needed.
		resp.RequiresReplace = false
		return
	}

	resp.RequiresReplace = true
}

func networksNestedValueRequiresReplaceIf() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			var plan, state loadBalancerModel
			resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

			if resp.Diagnostics.HasError() {
				return
			}

			replaced, diags := networkReplacedWithNetworks(ctx, state, plan)
			resp.Diagnostics.Append(diags...)

			// Do not require replace when value changes from null to non-null value when `network` is replaced with `networks`
			if replaced && req.StateValue.ValueString() == "" && req.PlanValue.ValueString() != "" {
				resp.RequiresReplace = false
				return
			} else {
				resp.RequiresReplace = true
			}
		},
		networksRequiresReplaceIfDescription,
		networksRequiresReplaceIfDescription,
	)
}
