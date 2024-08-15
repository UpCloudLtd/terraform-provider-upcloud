package storage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type storageCommonModel struct {
	Encrypt      types.Bool   `tfsdk:"encrypt"`
	ID           types.String `tfsdk:"id"`
	Labels       types.Map    `tfsdk:"labels"`
	SystemLabels types.Map    `tfsdk:"system_labels"`
	Size         types.Int64  `tfsdk:"size"`
	Tier         types.String `tfsdk:"tier"`
	Title        types.String `tfsdk:"title"`
	Type         types.String `tfsdk:"type"`
	Zone         types.String `tfsdk:"zone"`
}

func commonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"encrypt": schema.BoolAttribute{
			MarkdownDescription: "Sets if the storage is encrypted at rest.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplace(),
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
			Required:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 4096),
			},
		},
		"tier": schema.StringAttribute{
			MarkdownDescription: "The tier of the storage.",
			Computed:            true,
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					upcloud.StorageTierMaxIOPS,
					upcloud.StorageTierStandard,
					upcloud.StorageTierHDD,
				),
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
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func setCommonValues(ctx context.Context, data *storageCommonModel, storage *upcloud.StorageDetails) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.ID = types.StringValue(storage.UUID)
	data.Encrypt = types.BoolValue(storage.Encrypted.Bool())
	data.Size = types.Int64Value(int64(storage.Size))
	data.Tier = types.StringValue(storage.Tier)
	data.Title = types.StringValue(storage.Title)
	data.Type = types.StringValue(storage.Type)
	data.Zone = types.StringValue(storage.Zone)

	var diags diag.Diagnostics
	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(storage.Labels))
	respDiagnostics.Append(diags...)

	data.SystemLabels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToSystemLabelsMap(storage.Labels))
	respDiagnostics.Append(diags...)

	return respDiagnostics
}
