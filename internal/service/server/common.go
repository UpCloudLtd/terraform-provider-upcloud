package server

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverCommonModel struct {
	BootOrder         types.String `tfsdk:"boot_order"`
	CPU               types.Int64  `tfsdk:"cpu"`
	Firewall          types.Bool   `tfsdk:"firewall"`
	Host              types.Int64  `tfsdk:"host"`
	Hostname          types.String `tfsdk:"hostname"`
	ID                types.String `tfsdk:"id"`
	Labels            types.Map    `tfsdk:"labels"`
	Mem               types.Int64  `tfsdk:"mem"`
	Metadata          types.Bool   `tfsdk:"metadata"`
	NetworkInterfaces types.List   `tfsdk:"network_interface"`
	NICModel          types.String `tfsdk:"nic_model"`
	Plan              types.String `tfsdk:"plan"`
	ServerGroup       types.String `tfsdk:"server_group"`
	Tags              types.Set    `tfsdk:"tags"`
	Timezone          types.String `tfsdk:"timezone"`
	Title             types.String `tfsdk:"title"`
	VideoModel        types.String `tfsdk:"video_model"`
	Zone              types.String `tfsdk:"zone"`
}

func setCommonValues(ctx context.Context, data *serverCommonModel, details *upcloud.ServerDetails) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.BootOrder = types.StringValue(details.BootOrder)
	data.CPU = types.Int64Value(int64(details.CoreNumber))
	data.Firewall = types.BoolValue(details.Firewall == "on")
	data.Host = types.Int64Value(details.HostID)
	data.Hostname = types.StringValue(details.Hostname)
	data.Mem = types.Int64Value(int64(details.MemoryAmount))
	data.NICModel = types.StringValue(details.NICModel)
	data.Plan = types.StringValue(details.Plan)
	data.Timezone = types.StringValue(details.Timezone)
	data.Title = types.StringValue(details.Title)
	data.VideoModel = types.StringValue(details.VideoModel)
	data.Zone = types.StringValue(details.Zone)

	// Only set these fields if they have been configured or when they are undefined (i.e. called from data-source).
	if !data.ServerGroup.IsNull() {
		data.ServerGroup = types.StringValue(details.ServerGroup)
	}
	if !data.Metadata.IsNull() {
		data.Metadata = types.BoolValue(details.Metadata.Bool())
	}

	var diags diag.Diagnostics

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(details.Labels))
	respDiagnostics.Append(diags...)

	return respDiagnostics
}
