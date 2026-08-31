package loadbalancer

import (
	"context"
	"time"

	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type certificateBundleCommonModel struct {
	ID               types.String `tfsdk:"id"`
	Labels           types.Map    `tfsdk:"labels"`
	Name             types.String `tfsdk:"name"`
	NotAfter         types.String `tfsdk:"not_after"`
	NotBefore        types.String `tfsdk:"not_before"`
	OperationalState types.String `tfsdk:"operational_state"`
}

func setCertificateBundleCommonValues(ctx context.Context, data *certificateBundleCommonModel, bundle *v9.LoadBalancerCertificateBundle) diag.Diagnostics {
	var diags, d diag.Diagnostics

	data.Name = types.StringValue(bundle.Name)
	data.NotAfter = types.StringValue(bundle.NotAfter.Format(time.RFC3339))
	data.NotBefore = types.StringValue(bundle.NotBefore.Format(time.RFC3339))
	data.OperationalState = types.StringValue(string(bundle.OperationalState))

	labelsMap := make(map[string]string)
	if bundle.Labels != nil {
		labelsMap = labelsV9SliceToMap(*bundle.Labels)
	}
	data.Labels, d = types.MapValueFrom(ctx, types.StringType, labelsMap)
	diags.Append(d...)

	return diags
}
