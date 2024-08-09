package storage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &storageTemplateResource{}
	_ resource.ResourceWithConfigure   = &storageTemplateResource{}
	_ resource.ResourceWithImportState = &storageTemplateResource{}
)

func NewStorageTemplateResource() resource.Resource {
	return &storageTemplateResource{}
}

type storageTemplateResource struct {
	client *service.Service
}

func (r *storageTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_template"
}

// Configure adds the provider configured client to the resource.
func (r *storageTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type storageTemplateModel struct {
	storageCommonModel

	SourceStorage types.String `tfsdk:"source_storage"`
}

func (r *storageTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UpCloud storage templates.",
		Attributes: map[string]schema.Attribute{
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: "Sets if the storage is encrypted at rest.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the storage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels":        utils.LabelsAttribute("storage"),
			"system_labels": utils.SystemLabelsAttribute("storage"),
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the storage in gigabytes.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"source_storage": schema.StringAttribute{
				MarkdownDescription: "The source storage that is used as a base for this storage template.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "The tier of the storage.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A short, informative description.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the storage.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "The zone the storage is in, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *storageTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data storageTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labelsMap map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelsMap, false)...)
	}
	labels := utils.NilAsEmptyList(utils.LabelsMapToSlice(labelsMap))

	storage, diags := templatizeStorage(ctx, r.client, request.TemplatizeStorageRequest{
		Title: data.Title.ValueString(),
		UUID:  data.SourceStorage.ValueString(),
	}, request.ModifyStorageRequest{
		Labels: &labels,
	})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, storage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data storageTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	storage, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read storage details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, storage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state storageTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	_, err := r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         data.ID.ValueString(),
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	apiReq := request.ModifyStorageRequest{
		Labels: &labelsSlice,
		Title:  data.Title.ValueString(),
		UUID:   data.ID.ValueString(),
	}

	storage, err := r.client.ModifyStorage(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setCommonValues(ctx, &data.storageCommonModel, storage)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *storageTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data storageTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	id := data.ID.ValueString()

	// Wait for storage to enter 'online' state as storage devices can only be deleted in this state.
	_, err := r.client.WaitForStorageState(ctx, &request.WaitForStorageStateRequest{
		UUID:         id,
		DesiredState: upcloud.StorageStateOnline,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage did not reach online state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	deleteStorageRequest := &request.DeleteStorageRequest{
		UUID: id,
	}
	err = r.client.DeleteStorage(ctx, deleteStorageRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete storage",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *storageTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
