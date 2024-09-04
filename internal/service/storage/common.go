package storage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	encryptDescription = "Sets if the storage is encrypted at rest."
	sizeDescription    = "The size of the storage in gigabytes."
	tierDescription    = "The tier of the storage."
	titleDescription   = "The title of the storage."
	typeDescription    = "The type of the storage."
	uuidDescription    = "UUID of the storage."
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

func setCommonValues(ctx context.Context, data *storageCommonModel, storage *upcloud.Storage) diag.Diagnostics {
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
