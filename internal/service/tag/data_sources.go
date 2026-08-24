package tag

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

func NewTagsDataSource() datasource.DataSource {
	return &tagsDataSource{}
}

var (
	_ datasource.DataSource              = &tagsDataSource{}
	_ datasource.DataSourceWithConfigure = &tagsDataSource{}
)

type tagsDataSource struct {
	client *service.Service
}

func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type tagsModel struct {
	ID   types.String `tfsdk:"id"`
	Tags types.Set    `tfsdk:"tags"`
}

func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `~> Consider using labels instead of tags. Tags are an access control feature and only available for a limited set of resources. Use labels to describe and filter your resources.

List tags configured in the current account.`, Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"tags": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: nameDescription,
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: descriptionDescription,
							Computed:    true,
						},
						"servers": schema.SetAttribute{
							Description: serversDescription,
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *tagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tagsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	data.ID = types.StringValue(time.Now().UTC().String())

	tags, err := d.client.GetTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read tags",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	tagsData := make([]tagCommonModel, 0, len(tags.Tags))
	for _, tag := range tags.Tags {
		t := tagCommonModel{}
		t.Servers = types.SetUnknown(types.StringType)
		resp.Diagnostics.Append(setValues(ctx, &t, &tag)...)
		tagsData = append(tagsData, t)
	}

	var diags diag.Diagnostics
	data.Tags, diags = types.SetValueFrom(ctx, data.Tags.ElementType(ctx), tagsData)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
