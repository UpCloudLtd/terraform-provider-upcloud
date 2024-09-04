package storage

import (
	"context"
	"regexp"
	"sort"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewStorageDataSource() datasource.DataSource {
	return &storageDataSource{}
}

var (
	_ datasource.DataSource                     = &storageDataSource{}
	_ datasource.DataSourceWithConfigure        = &storageDataSource{}
	_ datasource.DataSourceWithConfigValidators = &storageDataSource{}
)

type storageDataSource struct {
	client *service.Service
}

func (d *storageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (d *storageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type storageDataSourceModel struct {
	storageCommonModel

	AccessType types.String `tfsdk:"access_type"`
	Name       types.String `tfsdk:"name"`
	NameRegex  types.String `tfsdk:"name_regex"`
	State      types.String `tfsdk:"state"`
	MostRecent types.Bool   `tfsdk:"most_recent"`
}

const storageDataSourceDescription = `Provides information on UpCloud [Block Storage](https://upcloud.com/products/block-storage) devices.

Data source can be used to map storage to other resource based on the ID or just to read some other storage property like zone information. Storage types are: ` + "`" + `normal` + "`" + `, ` + "`" + `backup` + "`" + `, ` + "`" + `cdrom` + "`" + `, and ` + "`" + `template` + "`" + `.`

func (d *storageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: storageDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"access_type": schema.StringAttribute{
				MarkdownDescription: "The access type of the storage, `public` or `private`.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						upcloud.StorageAccessPublic,
						upcloud.StorageAccessPrivate,
					),
				},
			},
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: encryptDescription,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: uuidDescription,
			},
			"labels":        utils.ReadOnlyLabelsAttribute("storage"),
			"system_labels": utils.SystemLabelsAttribute("storage"),
			"size": schema.Int64Attribute{
				MarkdownDescription: sizeDescription,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				Description: "Current state of the storage",
				Computed:    true,
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: tierDescription,
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: titleDescription,
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: typeDescription,
				Computed:            true,
				Optional:            true,
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "The zone the storage is in, e.g. `de-fra1`.",
				Optional:            true,
				Computed:            true,
			},
			"name_regex": schema.StringAttribute{
				Description:        "Use regular expression to match storage name. Deprecated, use exact title or UUID instead.",
				DeprecationMessage: "Use exact title or UUID instead.",
				Optional:           true,
			},
			"name": schema.StringAttribute{
				Description:        "Exact name of the storage (same as title). Deprecated, use `title` instead.",
				DeprecationMessage: "Contains the same value as `title`. Use `title` instead.",
				Optional:           true,
			},
			"most_recent": schema.BoolAttribute{
				Description:        "If more than one result is returned, use the most recent storage. This is only useful with private storages. Public storages might give unpredictable results.",
				DeprecationMessage: "Use exact title or UUID to limit the number of matching storages. Note that if you have multiple storages with the same title, you should use UUID to select the storage.",
				Optional:           true,
			},
		},
		Blocks: map[string]schema.Block{},
	}
}

func (d *storageDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("name_regex"),
			path.MatchRoot("title"),
		),
	}
}

func (d *storageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var re *regexp.Regexp
	if !data.NameRegex.IsNull() {
		var err error
		re, err = regexp.Compile(data.NameRegex.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not compile name_regex into a regular expression",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	storageType := data.Type.ValueString()
	storages, err := d.client.GetStorages(ctx, &request.GetStoragesRequest{Type: storageType})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read storages",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	matches := make([]upcloud.Storage, 0)
	for _, storage := range storages.Storages {
		if !data.Zone.IsNull() && data.Zone.ValueString() != storage.Zone {
			continue
		}

		if !data.AccessType.IsNull() && data.AccessType.ValueString() != storage.Access {
			continue
		}

		if !data.ID.IsNull() && data.ID.ValueString() != storage.UUID {
			continue
		}

		if !data.Title.IsNull() && data.Title.ValueString() != storage.Title {
			continue
		}

		if !data.Name.IsNull() && data.Name.ValueString() != storage.Title {
			continue
		}

		if !data.NameRegex.IsNull() && re != nil && !re.MatchString(storage.Title) {
			continue
		}

		matches = append(matches, storage)
	}

	if len(matches) < 1 {
		resp.Diagnostics.AddError("query returned no results", "")
		return
	}

	if len(matches) > 1 {
		if !data.MostRecent.ValueBool() {
			resp.Diagnostics.AddError("query returned more than one result", "")
			return
		}

		hasUnpredictableResults := false
		if !data.AccessType.IsNull() && data.AccessType.ValueString() == upcloud.StorageAccessPublic {
			// sort storages by UUID because public templates are missing 'created' timestamp
			hasUnpredictableResults = true
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].UUID > matches[j].UUID
			})
		} else {
			// sort storages by created timestamp
			sort.Slice(matches, func(i, j int) bool {
				if !hasUnpredictableResults && (matches[i].Created.IsZero() || matches[j].Created.IsZero()) {
					hasUnpredictableResults = true
				}
				return matches[i].Created.Unix() > matches[j].Created.Unix()
			})
		}
		if hasUnpredictableResults {
			resp.Diagnostics.AddWarning("using 'most_recent' attribute with public images might give unpredictable results", "")
		}
	}

	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, &matches[0])...)
	data.State = types.StringValue(matches[0].State)
	data.AccessType = types.StringValue(matches[0].Access)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
