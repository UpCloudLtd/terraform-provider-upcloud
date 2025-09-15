package ip

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewIPAddressesDataSource() datasource.DataSource {
	return &ipAddressesDataSource{}
}

var (
	_ datasource.DataSource              = &ipAddressesDataSource{}
	_ datasource.DataSourceWithConfigure = &ipAddressesDataSource{}
)

type ipAddressesDataSource struct {
	client *service.Service
}

func (d *ipAddressesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_addresses"
}

func (d *ipAddressesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type ipAddressesDataSourceModel struct {
	Addresses types.Set    `tfsdk:"addresses"`
	ID        types.String `tfsdk:"id"`
}

type ipAddressesAddressesModel struct {
	Access        types.String `tfsdk:"access"`
	Address       types.String `tfsdk:"address"`
	Family        types.String `tfsdk:"family"`
	Floating      types.Bool   `tfsdk:"floating"`
	MAC           types.String `tfsdk:"mac"`
	PartOfPlan    types.Bool   `tfsdk:"part_of_plan"`
	PTRRecord     types.String `tfsdk:"ptr_record"`
	ReleasePolicy types.String `tfsdk:"release_policy"`
	Server        types.String `tfsdk:"server"`
	Zone          types.String `tfsdk:"zone"`
}

func (d *ipAddressesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns a set of IP Addresses that are associated with the UpCloud account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the resource.",
			},
		},
		Blocks: map[string]schema.Block{
			"addresses": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Is address for utility or public network",
						},
						"address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "An UpCloud assigned IP Address",
						},
						"family": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "IP address family",
						},
						"floating": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Does the IP Address represents a floating IP Address",
						},
						"mac": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "MAC address of server interface to assign address to",
						},
						"part_of_plan": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Is the address a part of a plan",
						},
						"ptr_record": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A reverse DNS record entry",
						},
						"release_policy": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Release policy for the address",
						},
						"server": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for a server",
						},
						"zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Zone of address, required when assigning a detached floating IP address, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
						},
					},
				},
			},
		},
	}
}

func (d *ipAddressesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data ipAddressesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	ipAddresses, err := d.client.GetIPAddresses(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read ip addresses",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	dataIPAddresses := make([]ipAddressesAddressesModel, 0)
	for _, ipAddress := range ipAddresses.IPAddresses {
		var address ipAddressesAddressesModel
		address.Access = types.StringValue(ipAddress.Access)
		address.Address = types.StringValue(ipAddress.Address)
		address.Family = types.StringValue(ipAddress.Family)
		address.Floating = types.BoolValue(ipAddress.Floating.Bool())
		address.MAC = types.StringValue(ipAddress.MAC)
		address.PartOfPlan = types.BoolValue(ipAddress.PartOfPlan.Bool())
		address.PTRRecord = types.StringValue(ipAddress.PTRRecord)
		address.ReleasePolicy = types.StringValue(string(ipAddress.ReleasePolicy))
		address.Server = types.StringValue(ipAddress.ServerUUID)
		address.Zone = types.StringValue(ipAddress.Zone)

		dataIPAddresses = append(dataIPAddresses, address)
	}

	data.Addresses, diags = types.SetValueFrom(ctx, data.Addresses.ElementType(ctx), dataIPAddresses)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
