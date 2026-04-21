package managedobjectstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStorageBucketResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageBucketResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageBucketResource{}
)

func NewBucketResource() resource.Resource {
	return &managedObjectStorageBucketResource{}
}

type managedObjectStorageBucketResource struct {
	client *v9.ClientWithResponses
}

func (r *managedObjectStorageBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_bucket"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type bucketModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceUUID    types.String `tfsdk:"service_uuid"`
	Name           types.String `tfsdk:"name"`
	TotalObjects   types.Int64  `tfsdk:"total_objects"`
	TotalSizeBytes types.Int64  `tfsdk:"total_size_bytes"`
}

func (r *managedObjectStorageBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource represents an UpCloud Managed Object Storage bucket.

~> This resource uses the UpCloud API to manage the Managed Object Storage buckets. The main difference to S3 API is that the buckets can be deleted even when the bucket contains objects.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the bucket.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the bucket. ID is in {object storage UUID}/{bucket name} format.",
				Computed:    true,
			},
			"service_uuid": schema.StringAttribute{
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"total_objects": schema.Int64Attribute{
				MarkdownDescription: "Number of objects stored in the bucket.",
				Computed:            true,
			},
			"total_size_bytes": schema.Int64Attribute{
				MarkdownDescription: "Total size of objects stored in the bucket.",
				Computed:            true,
			},
		},
	}
}

func setBucketValues(data *bucketModel, detail *v9.ObjectStorage2BucketDetailResponse, fallbackName string) {
	if detail == nil {
		return
	}
	if detail.Name != nil {
		data.Name = types.StringValue(*detail.Name)
	} else if fallbackName != "" {
		data.Name = types.StringValue(fallbackName)
	}
	var totalObj int64
	if detail.TotalObjects != nil {
		totalObj = int64(*detail.TotalObjects)
	}
	data.TotalObjects = types.Int64Value(totalObj)
	var totalSize int64
	if detail.TotalSizeBytes != nil {
		totalSize = *detail.TotalSizeBytes
	}
	data.TotalSizeBytes = types.Int64Value(totalSize)
}

func objectStorageAPIErrorDetail(prob *v9.ObjectStorage2ErrorResponse, body []byte) string {
	if prob != nil {
		return fmt.Sprintf("%s (HTTP %d, type=%s, correlation_id=%s)", prob.Title, prob.Status, prob.Type, prob.CorrelationId)
	}
	if len(body) > 0 {
		return string(body)
	}
	return "unexpected API response"
}

func (r *managedObjectStorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Name.ValueString()))

	svcUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.CreateObjectStorageBucketWithResponse(ctx, svcUUID, v9.CreateObjectStorageBucketJSONRequestBody{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage bucket",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage bucket",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
		return
	}
	if apiResp.JSON201 == nil {
		var dest v9.ObjectStorage2CreateBucket201
		if err := json.Unmarshal(apiResp.Body, &dest); err != nil {
			resp.Diagnostics.AddError(
				"Unable to read created managed object storage bucket",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		apiResp.JSON201 = &dest
	}

	setBucketValues(&data, apiResp.JSON201, data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getBucket(ctx context.Context, serviceUUID, name string, client *v9.ClientWithResponses) (*v9.ObjectStorage2BucketDetailResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		diags.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	apiResp, err := client.ListObjectStorageBucketMetricsWithResponse(ctx, svcUUID, nil)
	if err != nil {
		diags.AddError(
			"Unable to read managed object storage buckets",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		return nil, diags
	}
	if apiResp.StatusCode() != http.StatusOK {
		diags.AddError(
			"Unable to read managed object storage buckets",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
		return nil, diags
	}
	if apiResp.JSON200 == nil {
		var dest v9.ObjectStorage2ListBucketMetrics200
		if err := json.Unmarshal(apiResp.Body, &dest); err != nil {
			diags.AddError(
				"Unable to read managed object storage buckets",
				utils.ErrorDiagnosticDetail(err),
			)
			return nil, diags
		}
		apiResp.JSON200 = &dest
	}
	buckets := *apiResp.JSON200
	for i := range buckets {
		b := buckets[i]
		if b.Name != nil && *b.Name == name {
			return &b, diags
		}
	}
	return nil, diags
}

func (r *managedObjectStorageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUID, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	bucket, diags := getBucket(ctx, serviceUUID, name, r.client)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if bucket == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	setBucketValues(&data, bucket, name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state bucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUUID, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, diags := getBucket(ctx, serviceUUID, name, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if bucket == nil {
		resp.Diagnostics.AddError(
			"Bucket not found",
			"Bucket with given name not found from the service",
		)
		return
	}

	setBucketValues(&data, bucket, name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &name)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	apiResp, err := r.client.DeleteObjectStorageBucketWithResponse(ctx, svcUUID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage bucket",
			utils.ErrorDiagnosticDetail(err),
		)
	} else if apiResp.StatusCode() != http.StatusNoContent && apiResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage bucket",
			objectStorageAPIErrorDetail(apiResp.ApplicationproblemJSONDefault, apiResp.Body),
		)
	}

	if err := r.client.WaitForObjectStorageBucketDeletion(ctx, serviceUUID, name); err != nil {
		resp.Diagnostics.AddError(
			"The deleted bucket was not removed from the managed object storage service",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
