package server

import (
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
