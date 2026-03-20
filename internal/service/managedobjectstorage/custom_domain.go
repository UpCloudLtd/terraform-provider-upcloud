package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api-generated/pkg/upcloud"
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
	// TODO: add mode when it is added to OpenAPI spec
	// Mode        types.String `tfsdk:"mode"`
	Type types.String `tfsdk:"type"`
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
			// TODO: add mode when it is added to OpenAPI spec
			// "mode": schema.StringAttribute{
			// 	MarkdownDescription: "Routing mode for the domain. Defaults to `api`.",
			// 	Optional:            true,
			// 	Computed:            true,
			// 	Default:             stringdefault.StaticString("api"),
			// 	Validators: []validator.String{
			// 		stringvalidator.OneOf("api", "static-website"),
			// 	},
			// },
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

	uuid, err := uuid.Parse(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	_, err = r.client.AttachObjectStorageCustomDomainWithResponse(
		ctx,
		uuid,
		v9.AttachObjectStorageCustomDomainJSONRequestBody{
			DomainName: data.DomainName.ValueString(),
			Type:       v9.ObjectStorage2CustomDomainCreateType(data.Type.ValueString()),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
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

	var serviceUUID, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceUUID = types.StringValue(serviceUUID)
	uuid, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	customDomain, err := r.client.GetObjectStorageCustomDomainWithResponse(
		ctx,
		uuid,
		domainName,
	)

	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read managed object storage custom domain details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	data.DomainName = types.StringValue(*customDomain.JSON200.DomainName)
	data.Type = types.StringValue(*customDomain.JSON200.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageCustomDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state customDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUUID, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &serviceUUID, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	body := v9.ModifyObjectStorageCustomDomainJSONRequestBody{
		DomainName: data.DomainName.ValueStringPointer(),
	}

	if data.Type.ValueString() != "" {
		t := v9.ObjectStorage2CustomDomainModifyType(data.Type.ValueString())
		body.Type = &t
	}

	customDomain, err := r.client.ModifyObjectStorageCustomDomainWithResponse(ctx,
		uuid,
		domainName,
		body,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))
	data.DomainName = types.StringPointerValue(customDomain.JSON200.DomainName)
	data.Type = types.StringPointerValue(customDomain.JSON200.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageCustomDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uuid, err := uuid.Parse(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	_, err = r.client.DeleteObjectStorageCustomDomainWithResponse(
		ctx,
		uuid,
		domainName,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageCustomDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
