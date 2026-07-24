package server

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

type serverDataSource struct {
	client *service.Service
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type serverDataSourceModel struct {
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
	State             types.String `tfsdk:"state"`
	Tags              types.Set    `tfsdk:"tags"`
	Timezone          types.String `tfsdk:"timezone"`
	Title             types.String `tfsdk:"title"`
	VideoModel        types.String `tfsdk:"video_model"`
	Zone              types.String `tfsdk:"zone"`
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	additionalIPAddressAttrs := map[string]schema.Attribute{
		"ip_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "An additional IP address for this interface.",
		},
		"ip_address_family": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The type of the additional IP address (`IPv4` or `IPv6`).",
		},
		"ip_address_floating": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "`true` indicates that the additional IP address is a floating IP address.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides details about an existing [UpCloud cloud server](https://upcloud.com/products/cloud-servers) by UUID.",
		Attributes: map[string]schema.Attribute{
			"boot_order": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The boot device order, `cdrom`|`disk`|`network` or comma separated combination of those values.",
			},
			"cpu": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of CPU cores for the server.",
			},
			"firewall": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Are firewall rules active for the server.",
			},
			"host": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The host the server is running on. Only available for private cloud hosts.",
			},
			"hostname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The hostname of the server.",
			},
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the server.",
			},
			"labels": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "User defined key-value pairs to classify the server.",
			},
			"mem": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The amount of memory for the server (in megabytes).",
			},
			"metadata": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Is the metadata service active for the server.",
			},
			"network_interface": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The network interfaces attached to the server.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"additional_ip_address": schema.SetNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Additional IP addresses for this interface. Only available for private network interfaces.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: additionalIPAddressAttrs,
							},
						},
						"bootable": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "`true` indicates that the server can boot from this interface.",
						},
						"index": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The interface index.",
						},
						"ip_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The primary IP address of this interface.",
						},
						"ip_address_family": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The type of the primary IP address (`IPv4` or `IPv6`).",
						},
						"ip_address_floating": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "`true` indicates that the primary IP address is a floating IP address.",
						},
						"mac_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The MAC address of the interface.",
						},
						"network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The UUID of the network the interface is attached to.",
						},
						"source_ip_filtering": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "`true` indicates that source IP filtering is enabled for this interface.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The type of the network interface (`public`, `private`, or `utility`).",
						},
					},
				},
			},
			"nic_model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the server's network interfaces.",
			},
			"plan": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The pricing plan used for the server.",
			},
			"server_group": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the server group the server belongs to.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current state of the server (`started`, `stopped`, `maintenance`, `error`).",
			},
			"tags": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tags attached to the server.",
			},
			"timezone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timezone of the server.",
			},
			"title": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A short, informational description of the server.",
			},
			"video_model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the server's video interface.",
			},
			"zone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The zone in which the server is hosted, e.g. `de-fra1`.",
			},
		},
	}
}

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := d.client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read server details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.BootOrder = types.StringValue(details.BootOrder)
	data.CPU = types.Int64Value(int64(details.CoreNumber))
	data.Firewall = types.BoolValue(details.Firewall == "on")
	data.Host = types.Int64Value(details.HostID)
	data.Hostname = types.StringValue(details.Hostname)
	data.Mem = types.Int64Value(int64(details.MemoryAmount))
	data.Metadata = types.BoolValue(details.Metadata.Bool())
	data.NICModel = types.StringValue(details.NICModel)
	data.Plan = types.StringValue(details.Plan)
	data.ServerGroup = types.StringValue(details.ServerGroup)
	data.State = types.StringValue(details.State)
	data.Timezone = types.StringValue(details.Timezone)
	data.Title = types.StringValue(details.Title)
	data.VideoModel = types.StringValue(details.VideoModel)
	data.Zone = types.StringValue(details.Zone)

	var diags diag.Diagnostics

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(details.Labels))
	resp.Diagnostics.Append(diags...)

	data.Tags, diags = types.SetValueFrom(ctx, types.StringType, []string(details.Tags))
	resp.Diagnostics.Append(diags...)

	networkInterfaces := make([]networkInterfaceModel, 0, len(details.Networking.Interfaces))
	for _, iface := range details.Networking.Interfaces {
		ni, d := setInterfaceValues(ctx, (*upcloud.Interface)(&iface), types.StringNull())
		resp.Diagnostics.Append(d...)
		networkInterfaces = append(networkInterfaces, ni)
	}

	data.NetworkInterfaces, diags = types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: networkInterfaceModel{}.AttributeTypes()},
		networkInterfaces,
	)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
