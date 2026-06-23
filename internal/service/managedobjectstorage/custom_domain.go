package managedobjectstorage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &managedObjectStorageCustomDomainResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageCustomDomainResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageCustomDomainResource{}
)

func NewCustomDomainResource() resource.Resource {
	return &managedObjectStorageCustomDomainResource{}
}

type managedObjectStorageCustomDomainResource struct {
	client *v9.ClientWithResponses
}

func (r *managedObjectStorageCustomDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_custom_domain"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageCustomDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type customDomainModel struct {
	DomainName  types.String `tfsdk:"domain_name"`
	ID          types.String `tfsdk:"id"`
	ServiceUUID types.String `tfsdk:"service_uuid"`
	Type        types.String `tfsdk:"type"`
}

func (r *managedObjectStorageCustomDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage custom domain. Note that DNS settings for the custom domain should be configured before creating this resource.",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				Description: "Must be a subdomain and consist of 3 to 5 parts such as objects.example.com. Cannot be root-level domain e.g. example.com.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "ID of the custom domain. ID is in {object storage UUID}/{domain name} format.",
				Computed:    true,
			},
			"service_uuid": schema.StringAttribute{
				Description: "Managed Object Storage service UUID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "At the moment only `public` is accepted.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("public"),
				Validators: []validator.String{
					stringvalidator.OneOf("public"),
				},
			},
		},
	}
}

func (r *managedObjectStorageCustomDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data customDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))

	serviceUUID, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.AttachObjectStorageCustomDomainWithResponse(ctx, serviceUUID, v9.ObjectStorage2CustomDomainCreate{
		DomainName: data.DomainName.ValueString(),
		Type:       v9.ObjectStorage2CustomDomainCreateType(data.Type.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage custom domain",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageCustomDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data customDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	var serviceUUIDStr, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUIDStr, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, err := uuid.Parse(serviceUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUIDStr)
	apiResp, err := r.client.GetObjectStorageCustomDomainWithResponse(ctx, serviceUUID, domainName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage custom domain details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage custom domain details",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	customDomain := apiResp.JSON200
	if customDomain.DomainName != nil {
		data.DomainName = types.StringValue(*customDomain.DomainName)
	}
	if customDomain.Type != nil {
		data.Type = types.StringValue(*customDomain.Type)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageCustomDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state customDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUUIDStr, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &serviceUUIDStr, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, err := uuid.Parse(serviceUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	newDomainName := data.DomainName.ValueString()
	newType := v9.ObjectStorage2CustomDomainModifyType(data.Type.ValueString())
	apiResp, err := r.client.ModifyObjectStorageCustomDomainWithResponse(ctx, serviceUUID, domainName, v9.ObjectStorage2CustomDomainModify{
		DomainName: &newDomainName,
		Type:       &newType,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage custom domain",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
		return
	}

	customDomain := apiResp.JSON200
	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))
	if customDomain.DomainName != nil {
		data.DomainName = types.StringValue(*customDomain.DomainName)
	}
	if customDomain.Type != nil {
		data.Type = types.StringValue(*customDomain.Type)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageCustomDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUIDStr, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUIDStr, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceUUID, err := uuid.Parse(serviceUUIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service UUID", utils.ErrorDiagnosticDetail(err))
		return
	}

	apiResp, err := r.client.DeleteObjectStorageCustomDomainWithResponse(ctx, serviceUUID, domainName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if apiResp.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage custom domain",
			fmt.Sprintf("API returned unexpected status %s", apiResp.Status()),
		)
	}
}

func (r *managedObjectStorageCustomDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
