package managedobjectstorage

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStorageResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageResource{}
)

func NewManagedObjectStorageResource() resource.Resource {
	return &managedObjectStorageResource{}
}

type managedObjectStorageResource struct {
	client *service.Service
}

func (r *managedObjectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type managedObjectStorageModel struct {
	ConfiguredStatus types.String `tfsdk:"configured_status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	Endpoint         types.Set    `tfsdk:"endpoint"`
	ID               types.String `tfsdk:"id"`
	Labels           types.Map    `tfsdk:"labels"`
	Name             types.String `tfsdk:"name"`
	Network          types.Set    `tfsdk:"network"`
	OperationalState types.String `tfsdk:"operational_state"`
	Region           types.String `tfsdk:"region"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

type endpointModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	IAMURL     types.String `tfsdk:"iam_url"`
	STSURL     types.String `tfsdk:"sts_url"`
	Type       types.String `tfsdk:"type"`
}

type networkModel struct {
	Family types.String `tfsdk:"family"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	UUID   types.String `tfsdk:"uuid"`
}

func (r *managedObjectStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage instance, which provides S3 compatible storage.",
		Attributes: map[string]schema.Attribute{
			"configured_status": schema.StringAttribute{
				Description: "Service status managed by the end user.",
				Required:    true,
				Validators: []validator.String{stringvalidator.OneOf(
					string(upcloud.ManagedObjectStorageConfiguredStatusStarted),
					string(upcloud.ManagedObjectStorageConfiguredStatusStopped),
				)},
			},
			"created_at": schema.StringAttribute{
				Description: "Creation time.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": schema.SetNestedAttribute{
				Description: "Endpoints for accessing the Managed Object Storage service.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain_name": schema.StringAttribute{
							MarkdownDescription: "Domain name of the endpoint.",
							Computed:            true,
						},
						"iam_url": schema.StringAttribute{
							Description: "URL for IAM.",
							Computed:    true,
						},
						"sts_url": schema.StringAttribute{
							Description: "URL for STS.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the endpoint (`private` / `public`).",
							Computed:    true,
						},
					},
				},
			},
			"labels": utils.LabelsAttribute("managed object storage"),
			"name": schema.StringAttribute{
				Description: "Name of the Managed Object Storage service. Must be unique within account.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The UUID of the managed object storage instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"operational_state": schema.StringAttribute{
				Description: "Operational state of the Managed Object Storage service.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region in which the service will be hosted, see `upcloud_managed_object_storage_regions` data source or use `upctl object-storage regions` to list available regions.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Update time.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"network": schema.SetNestedBlock{
				MarkdownDescription: "Attached networks from where object storage can be used. Private networks must reside in object storage region. To gain access from multiple private networks that might reside in different zones, create the networks and a corresponding router for each network.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"family": schema.StringAttribute{
							Description: "Network family. Currently only `IPv4` is supported.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.LoadBalancerAddressFamilyIPv4),
								),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Network name. Must be unique within the service.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
								stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must only contain letters, numbers, hyphens and underscores"),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Network type (`private` or `public`).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.LoadBalancerNetworkTypePrivate),
									string(upcloud.LoadBalancerNetworkTypePublic),
								),
							},
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Private network uuid. For public networks the field should be omitted.",
							Optional:            true,
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(0, 8),
				},
			},
		},
	}
}

func setManagedObjectStorageValues(ctx context.Context, data *managedObjectStorageModel, objsto *upcloud.ManagedObjectStorage) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.Name.ValueString() == ""

	data.ConfiguredStatus = types.StringValue(string(objsto.ConfiguredStatus))
	data.CreatedAt = types.StringValue(objsto.CreatedAt.String())
	data.Name = types.StringValue(objsto.Name)
	data.OperationalState = types.StringValue(string(objsto.OperationalState))
	data.Region = types.StringValue(objsto.Region)
	data.UpdatedAt = types.StringValue(objsto.UpdatedAt.String())

	endpoints := make([]endpointModel, 0)
	for _, endpoint := range objsto.Endpoints {
		endpoints = append(endpoints, endpointModel{
			DomainName: types.StringValue(endpoint.DomainName),
			IAMURL:     types.StringValue(endpoint.IAMURL),
			STSURL:     types.StringValue(endpoint.STSURL),
			Type:       types.StringValue(endpoint.Type),
		})
	}

	data.Endpoint, diags = types.SetValueFrom(ctx, data.Endpoint.ElementType(ctx), endpoints)
	respDiagnostics.Append(diags...)

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(objsto.Labels))
	respDiagnostics.Append(diags...)

	if isImport || !data.Network.IsNull() {
		networks := make([]networkModel, 0)
		for _, network := range objsto.Networks {
			networks = append(networks, networkModel{
				Family: types.StringValue(network.Family),
				Name:   types.StringValue(network.Name),
				Type:   types.StringValue(network.Type),
				UUID:   types.StringPointerValue(network.UUID),
			})
		}

		data.Network, diags = types.SetValueFrom(ctx, data.Network.ElementType(ctx), networks)
		respDiagnostics.Append(diags...)
	}

	return respDiagnostics
}

func buildNetworks(ctx context.Context, dataNetworks types.Set) ([]upcloud.ManagedObjectStorageNetwork, diag.Diagnostics) {
	var planNetworks []networkModel
	respDiagnostics := dataNetworks.ElementsAs(ctx, &planNetworks, false)

	networks := make([]upcloud.ManagedObjectStorageNetwork, 0)
	for _, network := range planNetworks {
		var uuid *string
		if !network.UUID.IsNull() {
			u := network.UUID.ValueString()
			uuid = &u
		}

		networks = append(networks, upcloud.ManagedObjectStorageNetwork{
			Family: network.Family.ValueString(),
			Name:   network.Name.ValueString(),
			Type:   network.Type.ValueString(),
			UUID:   uuid,
		})
	}

	return networks, respDiagnostics
}

func (r *managedObjectStorageResource) waitForManagedObjectStorageState(ctx context.Context, objsto *upcloud.ManagedObjectStorage) (*upcloud.ManagedObjectStorage, diag.Diagnostics) {
	var diags diag.Diagnostics

	waitReq := &request.WaitForManagedObjectStorageOperationalStateRequest{
		DesiredState: upcloud.ManagedObjectStorageOperationalStateRunning,
		UUID:         objsto.UUID,
	}

	if objsto.ConfiguredStatus == upcloud.ManagedObjectStorageConfiguredStatusStopped {
		waitReq.DesiredState = upcloud.ManagedObjectStorageOperationalStateStopped
	}

	objsto, err := r.client.WaitForManagedObjectStorageOperationalState(ctx, waitReq)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error while waiting for managed object storage to be in %s state", waitReq.DesiredState),
			utils.ErrorDiagnosticDetail(err),
		)
	}

	return objsto, diags
}

func (r *managedObjectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	networks, diags := buildNetworks(ctx, data.Network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := &request.CreateManagedObjectStorageRequest{
		ConfiguredStatus: upcloud.ManagedObjectStorageConfiguredStatus(data.ConfiguredStatus.ValueString()),
		Name:             data.Name.ValueString(),
		Networks:         networks,
		Labels:           utils.LabelsMapToSlice(labels),
		Region:           data.Region.ValueString(),
	}

	objsto, err := r.client.CreateManagedObjectStorage(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(objsto.UUID)

	objsto, diags = r.waitForManagedObjectStorageState(ctx, objsto)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(setManagedObjectStorageValues(ctx, &data, objsto)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	objsto, err := r.client.GetManagedObjectStorage(ctx, &request.GetManagedObjectStorageRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setManagedObjectStorageValues(ctx, &data, objsto)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	networks, diags := buildNetworks(ctx, data.Network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := &request.ModifyManagedObjectStorageRequest{
		UUID:             data.ID.ValueString(),
		ConfiguredStatus: (*upcloud.ManagedObjectStorageConfiguredStatus)(data.ConfiguredStatus.ValueStringPointer()),
		Labels:           &labelsSlice,
		Name:             data.Name.ValueStringPointer(),
		Networks:         &networks,
	}

	objsto, err := r.client.ModifyManagedObjectStorage(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	objsto, diags = r.waitForManagedObjectStorageState(ctx, objsto)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(setManagedObjectStorageValues(ctx, &data, objsto)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteManagedObjectStorage(ctx, &request.DeleteManagedObjectStorageRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if err := r.client.WaitForManagedObjectStorageDeletion(ctx, &request.WaitForManagedObjectStorageDeletionRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for managed object storage to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
