package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
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

func NewManagedObjectStorageBucketResource() resource.Resource {
	return &managedObjectStorageBucketResource{}
}

type managedObjectStorageBucketResource struct {
	client *service.Service
}

func (r *managedObjectStorageBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_bucket"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
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

func setBucketValues(data *bucketModel, bucket *upcloud.ManagedObjectStorageBucketMetrics) {
	data.Name = types.StringValue(bucket.Name)
	data.TotalObjects = types.Int64Value(int64(bucket.TotalObjects))
	data.TotalSizeBytes = types.Int64Value(int64(bucket.TotalSizeBytes))
}

func (r *managedObjectStorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.Name.ValueString()))

	apiReq := &request.CreateManagedObjectStorageBucketRequest{
		Name:        data.Name.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	}

	bucket, err := r.client.CreateManagedObjectStorageBucket(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage bucket",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	setBucketValues(&data, &bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getBucket(ctx context.Context, serviceUUID, name string, client *service.Service) (bucket *upcloud.ManagedObjectStorageBucketMetrics, diags diag.Diagnostics) {
	buckets, err := client.GetManagedObjectStorageBucketMetrics(ctx, &request.GetManagedObjectStorageBucketMetricsRequest{
		ServiceUUID: serviceUUID,
	})
	if err != nil {
		diags.AddError(
			"Unable to read managed object storage buckets",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	for _, b := range buckets {
		if b.Name == name {
			bucket = &b
			break
		}
	}
	return bucket, diags
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

	serviceUUID, name, diags := unmarshalID(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	bucket, diags := getBucket(ctx, serviceUUID, name, r.client)
	resp.Diagnostics.Append(diags...)

	if bucket == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	setBucketValues(&data, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state bucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, domainName, diags := unmarshalID(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, diags := getBucket(ctx, serviceUUID, domainName, r.client)
	resp.Diagnostics.Append(diags...)
	if bucket == nil {
		resp.Diagnostics.AddError(
			"Bucket not found",
			"Bucket with given name not found from the service",
		)
		return
	}

	setBucketValues(&data, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	serviceUUID, name, diags := unmarshalID(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedObjectStorageBucket(ctx, &request.DeleteManagedObjectStorageBucketRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage bucket",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	if err := r.client.WaitForManagedObjectStorageBucketDeletion(ctx, &request.WaitForManagedObjectStorageBucketDeletionRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"The deleted bucket was not removed from the managed object storage service",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
