package cloud

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	allFilter     string = "all"
	publicFilter  string = "public"
	privateFilter string = "private"
)

func NewZonesDataSource() datasource.DataSource {
	return &zonesDataSource{}
}

var (
	_ datasource.DataSource              = &zonesDataSource{}
	_ datasource.DataSourceWithConfigure = &zonesDataSource{}
)

type zonesDataSource struct {
	client *service.Service
}

func (d *zonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zones"
}

func (d *zonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type zonesModel struct {
	ID         types.String `tfsdk:"id"`
	ZoneIDs    []string     `tfsdk:"zone_ids"`
	FilterType types.String `tfsdk:"filter_type"`
}

func (d *zonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Returns a list of available UpCloud zones.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"zone_ids": schema.SetAttribute{
				Computed:    true,
				Description: `List of zone IDs.`,
				ElementType: types.StringType,
			},
			"filter_type": schema.StringAttribute{
				Description: `Filter zones by type. Possible values are "all", "public" and "private". Default is "public".`,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(allFilter, publicFilter, privateFilter),
				},
			},
		},
	}
}

func (d *zonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data zonesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	zones, err := d.client.GetZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read zones",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	req.Config.GetAttribute(ctx, path.Root("filter_type"), &data.FilterType)
	if data.FilterType.ValueString() == "" {
		data.FilterType = types.StringValue(publicFilter)
	}

	data.ZoneIDs = utils.FilterZoneIDs(zones.Zones, func(zone upcloud.Zone) bool {
		switch data.FilterType.ValueString() {
		case privateFilter:
			return zone.Public != upcloud.True
		case publicFilter:
			return zone.Public == upcloud.True
		default:
			return true
		}
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
