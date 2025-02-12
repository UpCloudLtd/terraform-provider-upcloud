package cloud

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewHostsDataSource() datasource.DataSource {
	return &hostsDataSource{}
}

var (
	_ datasource.DataSource              = &hostsDataSource{}
	_ datasource.DataSourceWithConfigure = &hostsDataSource{}
)

type hostsDataSource struct {
	client *service.Service
}

func (d *hostsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts"
}

func (d *hostsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type hostsModel struct {
	ID    types.String `tfsdk:"id"`
	Hosts types.Set    `tfsdk:"hosts"`
}

type hostModel struct {
	ID             types.Int64  `tfsdk:"host_id"`
	Description    types.String `tfsdk:"description"`
	Zone           types.String `tfsdk:"zone"`
	WindowsEnabled types.Bool   `tfsdk:"windows_enabled"`
	Statistics     types.List   `tfsdk:"statistics"`
}

type statModel struct {
	Name      types.String  `tfsdk:"name"`
	Timestamp types.String  `tfsdk:"timestamp"`
	Value     types.Float64 `tfsdk:"value"`
}

func (m statModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":      types.StringType,
		"timestamp": types.StringType,
		"value":     types.Float64Type,
	}
}

func (d *hostsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Returns a list of available UpCloud hosts. 
		A host identifies the host server that virtual machines are run on. 
		Only hosts on private cloud to which the calling account has access to are available through this resource.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"hosts": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"host_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The unique id of the host",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Free form text describing the host",
						},
						"zone": schema.StringAttribute{
							Computed:    true,
							Description: "The zone the host is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
						},
						"windows_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "If true, this node can be used as a host for Windows servers.",
						},
					},
					Blocks: map[string]schema.Block{
						"statistics": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "The name of the statistic",
									},
									"timestamp": schema.StringAttribute{
										Computed:    true,
										Description: "The timestamp of the statistic",
									},
									"value": schema.Float64Attribute{
										Computed:    true,
										Description: "The value of the statistic",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *hostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data hostsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	hosts, err := d.client.GetHosts(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read hosts",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	dataHosts := make([]hostModel, len(hosts.Hosts))
	for i, host := range hosts.Hosts {
		dataHosts[i].ID = types.Int64Value(int64(host.ID))
		dataHosts[i].Description = types.StringValue(host.Description)
		dataHosts[i].Zone = types.StringValue(host.Zone)
		dataHosts[i].WindowsEnabled = utils.AsBool(host.WindowsEnabled)

		statistics := make([]statModel, len(host.Stats))
		for j, stat := range host.Stats {
			statistics[j].Name = types.StringValue(stat.Name)
			statistics[j].Timestamp = types.StringValue(stat.Timestamp.String())
			statistics[j].Value = types.Float64Value(stat.Value)
		}
		dataHosts[i].Statistics, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: statModel{}.AttributeTypes(),
		}, statistics)
		resp.Diagnostics.Append(diags...)
	}

	data.Hosts, diags = types.SetValueFrom(ctx, data.Hosts.ElementType(ctx), dataHosts)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
