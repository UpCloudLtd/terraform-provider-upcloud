package cloud

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewZoneDataSource() datasource.DataSource {
	return &zoneDataSource{}
}

var (
	_ datasource.DataSource              = &zoneDataSource{}
	_ datasource.DataSourceWithConfigure = &zoneDataSource{}
)

type zoneDataSource struct {
	client *service.Service
}

func (d *zoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *zoneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type zoneModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Public      types.Bool   `tfsdk:"public"`
	ParentZone  types.String `tfsdk:"parent_zone"`
}

func (d *zoneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides details on given zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the zone.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier of the zone. Contains the same value as `id`.",
			},
			"public": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the zone is a public zone.",
			},
			"parent_zone": schema.StringAttribute{
				Computed:    true,
				Description: "Public parent zone of an private cloud zone. Empty for public zones.",
			},
		},
	}
}

func (d *zoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data zoneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	id := data.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("`id` must be set", "`id` in `upcloud_zone` data source can't be empty.")
		return
	}

	zones, err := d.client.GetZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read zones",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	var zone upcloud.Zone
	for _, z := range zones.Zones {
		if z.ID == id {
			zone = z
			break
		}
	}

	if zone.ID == "" {
		resp.Diagnostics.AddError("Zone not found", "No zone found with the given ID.")
		return
	}

	data.Description = types.StringValue(zone.Description)
	data.Public = types.BoolValue(zone.Public.Bool())
	data.ParentZone = types.StringValue(zone.ParentZone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
