package loadbalancer

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
