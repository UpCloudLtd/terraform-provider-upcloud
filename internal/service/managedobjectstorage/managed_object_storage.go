package managedobjectstorage

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
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
	openapi_types "github.com/oapi-codegen/runtime/types"
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
	client *v9.ClientWithResponses
}

func (r *managedObjectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
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
					string(v9.ObjectStorage2PropertyConfiguredStatusStarted),
					string(v9.ObjectStorage2PropertyConfiguredStatusStopped),
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
									string(v9.ObjectStorage2NetworkFamilyIPv4),
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
									string(v9.ObjectStorage2NetworkTypePrivate),
									string(v9.ObjectStorage2NetworkTypePublic),
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

func labelsMapToV9Slice(m map[string]string) []v9.ObjectStorage2LabelCreate {
	labels := make([]v9.ObjectStorage2LabelCreate, 0, len(m))

	for k, v := range m {
		key := v9.ObjectStorage2LabelKey(k)
		value := v9.ObjectStorage2LabelValue(v)
		labels = append(labels, v9.ObjectStorage2LabelCreate{
			Key:   key,
			Value: &value,
		})
	}

	return labels
}

func labelsV9SliceToMap(labels []v9.ObjectStorage2LabelDetailResponse) map[string]string {
	result := make(map[string]string)

	for _, label := range labels {
		if label.Key == nil || label.Value == nil {
			continue
		}
		if strings.HasPrefix(*label.Key, "_") {
			continue
		}

		result[*label.Key] = *label.Value
	}

	return result
}

func setManagedObjectStorageValues(ctx context.Context, data *managedObjectStorageModel, objsto *v9.GetObjectStorage200) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.Name.ValueString() == ""

	data.ConfiguredStatus = types.StringNull()
	if objsto.ConfiguredStatus != nil {
		data.ConfiguredStatus = types.StringValue(string(*objsto.ConfiguredStatus))
	}

	data.CreatedAt = types.StringNull()
	if objsto.CreatedAt != nil {
		data.CreatedAt = types.StringValue(objsto.CreatedAt.String())
	}

	data.Name = types.StringPointerValue(objsto.Name)

	data.OperationalState = types.StringNull()
	if objsto.OperationalState != nil {
		data.OperationalState = types.StringValue(string(*objsto.OperationalState))
	}

	data.Region = types.StringPointerValue(objsto.Region)

	data.UpdatedAt = types.StringNull()
	if objsto.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(objsto.UpdatedAt.String())
	}

	endpoints := make([]endpointModel, 0)
	if objsto.Endpoints != nil {
		for _, endpoint := range *objsto.Endpoints {
			endpointType := types.StringNull()
			if endpoint.Type != nil {
				endpointType = types.StringValue(string(*endpoint.Type))
			}

			endpoints = append(endpoints, endpointModel{
				DomainName: types.StringPointerValue(endpoint.DomainName),
				IAMURL:     types.StringPointerValue(endpoint.IamUrl),
				STSURL:     types.StringPointerValue(endpoint.StsUrl),
				Type:       endpointType,
			})
		}
	}

	data.Endpoint, diags = types.SetValueFrom(ctx, data.Endpoint.ElementType(ctx), endpoints)
	respDiagnostics.Append(diags...)

	labelsMap := make(map[string]string)
	if objsto.Labels != nil {
		labelsMap = labelsV9SliceToMap(*objsto.Labels)
	}

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, labelsMap)
	respDiagnostics.Append(diags...)

	if isImport || !data.Network.IsNull() {
		networks := make([]networkModel, 0)
		if objsto.Networks != nil {
			for _, network := range *objsto.Networks {
				networkUUID := types.StringNull()
				if network.Uuid != nil {
					networkUUID = types.StringValue(network.Uuid.String())
				}

				networks = append(networks, networkModel{
					Family: types.StringPointerValue(network.Family),
					Name:   types.StringPointerValue(network.Name),
					Type:   types.StringPointerValue(network.Type),
					UUID:   networkUUID,
				})
			}
		}

		data.Network, diags = types.SetValueFrom(ctx, data.Network.ElementType(ctx), networks)
		respDiagnostics.Append(diags...)
	}

	return respDiagnostics
}

// networkKey returns a string that uniquely identifies a managed object storage network.
func networkKey(n v9.ObjectStorage2NetworkCreate) string {
	uuid := ""
	if n.Uuid != nil {
		uuid = n.Uuid.String()
	}

	return fmt.Sprintf("%s/%s/%s/%s", n.Family, n.Name, n.Type, uuid)
}

// hasRemovedNetworks reports whether any network in current is absent from desired.
func hasRemovedNetworks(current, desired []v9.ObjectStorage2NetworkCreate) bool {
	desiredKeys := make(map[string]bool, len(desired))
	for _, n := range desired {
		desiredKeys[networkKey(n)] = true
	}

	for _, n := range current {
		if _, ok := desiredKeys[networkKey(n)]; !ok {
			return true
		}
	}

	return false
}

// retainedNetworks returns the networks from current that also appear in desired.
func retainedNetworks(current, desired []v9.ObjectStorage2NetworkCreate) []v9.ObjectStorage2NetworkCreate {
	desiredKeys := make(map[string]bool, len(desired))
	for _, n := range desired {
		desiredKeys[networkKey(n)] = true
	}

	retained := make([]v9.ObjectStorage2NetworkCreate, 0)
	for _, n := range current {
		if _, ok := desiredKeys[networkKey(n)]; ok {
			retained = append(retained, n)
		}
	}

	return retained
}

func buildNetworks(ctx context.Context, dataNetworks types.Set) ([]v9.ObjectStorage2NetworkCreate, diag.Diagnostics) {
	var planNetworks []networkModel
	respDiagnostics := dataNetworks.ElementsAs(ctx, &planNetworks, false)

	networks := make([]v9.ObjectStorage2NetworkCreate, 0)
	for _, network := range planNetworks {
		var networkUUID *openapi_types.UUID
		if !network.UUID.IsNull() && !network.UUID.IsUnknown() {
			u, err := uuid.Parse(network.UUID.ValueString())
			if err != nil {
				respDiagnostics.AddError(
					"Unable to parse object storage network UUID",
					utils.ErrorDiagnosticDetail(err),
				)
				continue
			}

			uuidValue := openapi_types.UUID(u)
			networkUUID = &uuidValue
		}

		networks = append(networks, v9.ObjectStorage2NetworkCreate{
			Family: v9.ObjectStorage2NetworkFamily(network.Family.ValueString()),
			Name:   v9.ObjectStorage2Name(network.Name.ValueString()),
			Type:   v9.ObjectStorage2NetworkType(network.Type.ValueString()),
			Uuid:   networkUUID,
		})
	}

	return networks, respDiagnostics
}

func (r *managedObjectStorageResource) waitForManagedObjectStorageState(ctx context.Context, serviceUUID string, configuredStatus v9.ObjectStorage2PropertyConfiguredStatus) (*v9.GetObjectStorage200, diag.Diagnostics) {
	var diags diag.Diagnostics

	desiredState := string(v9.ObjectStorage2ServiceDetailResponseOperationalStateRunning)
	if configuredStatus == v9.ObjectStorage2PropertyConfiguredStatusStopped {
		desiredState = string(v9.ObjectStorage2ServiceDetailResponseOperationalStateStopped)
	}

	objsto, err := r.client.WaitForObjectStorageOperationalState(ctx, serviceUUID, desiredState)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error while waiting for managed object storage to be in %s state", desiredState),
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
	labelsSlice := labelsMapToV9Slice(labels)
	configuredStatus := v9.ObjectStorage2PropertyConfiguredStatus(data.ConfiguredStatus.ValueString())

	apiReq := v9.CreateObjectStorageJSONRequestBody{
		ConfiguredStatus: configuredStatus,
		Name:             data.Name.ValueString(),
		Region:           v9.ObjectStorage2Name(data.Region.ValueString()),
	}
	if len(networks) > 0 {
		apiReq.Networks = &networks
	}
	if len(labelsSlice) > 0 {
		apiReq.Labels = &labelsSlice
	}

	apiResp, err := r.client.CreateObjectStorageWithResponse(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}
	if apiResp.JSON201.Uuid == nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage",
			"Create response did not include service UUID.",
		)
		return
	}

	data.ID = types.StringValue(*apiResp.JSON201.Uuid)

	objsto, diags := r.waitForManagedObjectStorageState(ctx, data.ID.ValueString(), configuredStatus)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	objstoResp, err := r.client.GetObjectStorageWithResponse(ctx, openapi_types.UUID(serviceUUID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if objstoResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if objstoResp.StatusCode() != http.StatusOK || objstoResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage details",
			fmt.Sprintf("Unexpected API status code %d: %s", objstoResp.StatusCode(), string(objstoResp.Body)),
		)
		return
	}

	resp.Diagnostics.Append(setManagedObjectStorageValues(ctx, &data, objstoResp.JSON200)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var state managedObjectStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := labelsMapToV9Slice(labels)

	networks, diags := buildNetworks(ctx, data.Network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateNetworks, diags := buildNetworks(ctx, state.Network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	configuredStatus := v9.ObjectStorage2PropertyConfiguredStatus(data.ConfiguredStatus.ValueString())

	// The API returns 500 when private networks are swapped in a single PATCH: remove old
	// networks first and wait for a stable state before adding the new ones.
	if hasRemovedNetworks(stateNetworks, networks) {
		intermediateNetworks := retainedNetworks(stateNetworks, networks)
		intermReq := v9.ModifyObjectStorageJSONRequestBody{
			ConfiguredStatus: &configuredStatus,
			Labels:           &labelsSlice,
			Name:             data.Name.ValueStringPointer(),
			Networks:         &intermediateNetworks,
		}
		intermResp, err := r.client.ModifyObjectStorageWithResponse(ctx, openapi_types.UUID(serviceUUID), intermReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to modify managed object storage",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
		if intermResp.StatusCode() != http.StatusOK || intermResp.JSON200 == nil {
			resp.Diagnostics.AddError(
				"Unable to modify managed object storage",
				fmt.Sprintf("Unexpected API status code %d: %s", intermResp.StatusCode(), string(intermResp.Body)),
			)
			return
		}

		_, intermDiags := r.waitForManagedObjectStorageState(ctx, data.ID.ValueString(), configuredStatus)
		resp.Diagnostics.Append(intermDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	apiReq := v9.ModifyObjectStorageJSONRequestBody{
		ConfiguredStatus: &configuredStatus,
		Labels:           &labelsSlice,
		Name:             data.Name.ValueStringPointer(),
		Networks:         &networks,
	}

	objstoResp, err := r.client.ModifyObjectStorageWithResponse(ctx, openapi_types.UUID(serviceUUID), apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if objstoResp.StatusCode() != http.StatusOK || objstoResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage",
			fmt.Sprintf("Unexpected API status code %d: %s", objstoResp.StatusCode(), string(objstoResp.Body)),
		)
		return
	}

	objsto, diags := r.waitForManagedObjectStorageState(ctx, data.ID.ValueString(), configuredStatus)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setManagedObjectStorageValues(ctx, &data, objsto)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data managedObjectStorageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	deleteParams := &v9.DeleteObjectStorageParams{}
	deleteResp, err := r.client.DeleteObjectStorageWithResponse(ctx, openapi_types.UUID(serviceUUID), deleteParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if deleteResp.StatusCode() != http.StatusNoContent && deleteResp.StatusCode() != http.StatusAccepted && deleteResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage",
			fmt.Sprintf("Unexpected API status code %d: %s", deleteResp.StatusCode(), string(deleteResp.Body)),
		)
		return
	}

	if err := r.client.WaitForObjectStorageDeletion(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for managed object storage to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
