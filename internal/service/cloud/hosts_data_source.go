package cloud

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	Hosts []hostModel  `tfsdk:"hosts"`
}

type hostModel struct {
	ID          types.Int64  `tfsdk:"host_id"`
	Description types.String `tfsdk:"description"`
	Zone        types.String `tfsdk:"zone"`
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
					},
				},
			},
		},
	}
}

func (d *hostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	data.Hosts = make([]hostModel, len(hosts.Hosts))
	for i, host := range hosts.Hosts {
		data.Hosts[i].ID = types.Int64Value(int64(host.ID))
		data.Hosts[i].Description = types.StringValue(host.Description)
		data.Hosts[i].Zone = types.StringValue(host.Zone)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
