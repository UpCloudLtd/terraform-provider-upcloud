package managedobjectstorage

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

var (
	_ datasource.DataSource              = &regionsDataSource{}
	_ datasource.DataSourceWithConfigure = &regionsDataSource{}
)

type regionsDataSource struct {
	client *service.Service
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_regions"
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type regionsModel struct {
	ID      types.String  `tfsdk:"id"`
	Regions []regionModel `tfsdk:"regions"`
}

type regionModel struct {
	Name        types.String `tfsdk:"name"`
	PrimaryZone types.String `tfsdk:"primary_zone"`
	Zones       types.Set    `tfsdk:"zones"`
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns a list of available Managed Object Storage regions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"regions": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the region.",
							Computed:    true,
						},
						"primary_zone": schema.StringAttribute{
							Description: "Primary zone of the region.",
							Computed:    true,
						},
						"zones": schema.SetAttribute{
							Description: "List of zones in the region.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data regionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	data.ID = types.StringValue(time.Now().UTC().String())

	regions, err := d.client.GetManagedObjectStorageRegions(ctx, &request.GetManagedObjectStorageRegionsRequest{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object-storage regions",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.Regions = make([]regionModel, len(regions))
	for i, region := range regions {
		data.Regions[i].Name = types.StringValue(region.Name)
		data.Regions[i].PrimaryZone = types.StringValue(region.PrimaryZone)

		zonesSlice := make([]string, 0)
		for _, zone := range region.Zones {
			zonesSlice = append(zonesSlice, zone.Name)
		}

		zones, diags := types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(zonesSlice))
		resp.Diagnostics.Append(diags...)
		data.Regions[i].Zones = zones
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
