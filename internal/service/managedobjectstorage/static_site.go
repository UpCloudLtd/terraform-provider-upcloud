package managedobjectstorage

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api-generated/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

/* TODO: uncomment once required methods are implemented
var (
	_ resource.Resource                = &managedObjectStorageStaticSiteResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageStaticSiteResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageStaticSiteResource{}
)

func NewStaticSiteResource() resource.Resource {
	return &managedObjectStorageStaticSiteResource{}
}
*/

type managedObjectStorageStaticSiteResource struct {
	client *v9.ClientWithResponses
}

func (r *managedObjectStorageStaticSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object_storage_static_site"
}

// Configure adds the provider configured client to the resource.
func (r *managedObjectStorageStaticSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type staticSiteModel struct {
	DomainName    types.String `tfsdk:"domain_name"`
	ID            types.String `tfsdk:"id"`
	ServiceUUID   types.String `tfsdk:"service_uuid"`
	BucketName    types.String `tfsdk:"bucket_name"`
	BucketPrefix  types.String `tfsdk:"bucket_prefix"`
	IndexDocument types.String `tfsdk:"index_document"`
	SpaMode       types.Bool   `tfsdk:"spa_mode"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ErrorPages    types.List   `tfsdk:"error_pages"`
}

type errorPageModel struct {
	StatusCode       types.Int64  `tfsdk:"status_code"`
	StatusRangeStart types.Int64  `tfsdk:"status_range_start"`
	StatusRangeEnd   types.Int64  `tfsdk:"status_range_end"`
	ErrorDocument    types.String `tfsdk:"error_document"`
}

func (r *managedObjectStorageStaticSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an UpCloud Managed Object Storage static site.",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "A custom domain in `static-website` mode attached to the service.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"bucket_name": schema.StringAttribute{
				MarkdownDescription: "S3 bucket from which static content is served. Must have a public read policy.",
				Required:            true,
			},
			"bucket_prefix": schema.StringAttribute{
				MarkdownDescription: "Path prefix within the bucket. For example, `dist/` serves content from the `dist/` folder. Defaults to empty string (bucket root).",
				Optional:            true,
				Computed:            true,
			},
			"index_document": schema.StringAttribute{
				Description: "Name of the index document. Defaults to `index.html`.",
				Optional:    true,
				Computed:    true,
			},
			"spa_mode": schema.BoolAttribute{
				Description: "Whether to enable single-page application mode. In SPA mode, the index document is served for all paths, which is useful for client-side routing.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable or disable serving content on this domain. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"error_page": schema.ListNestedBlock{
				MarkdownDescription: "Custom error pages served when the storage backend returns an error status.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"status_code": schema.Int64Attribute{
							Description: "Exact HTTP status code to match. Mutually exclusive with status range.",
							Optional:    true,
						},
						"status_range_start": schema.Int64Attribute{
							Description: "Start of the status code range (inclusive).",
							Optional:    true,
						},
						"status_range_end": schema.Int64Attribute{
							Description: "End of the status code range (inclusive). Must be greater than start.",
							Optional:    true,
						},
						"error_document": schema.StringAttribute{
							Description: "Path to the error page document within the bucket.",
							Required:    true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 25),
				},
			},
		},
	}
}

// TODO: replace custom domain API calls with static site API calls when they are added to OpenAPI spec
/*
func (r *managedObjectStorageStaticSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data staticSiteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))

	apiReq := &request.CreateManagedObjectStorageCustomDomainRequest{
		DomainName:  data.DomainName.ValueString(),
		Type:        data.Type.ValueString(),
		ServiceUUID: data.ServiceUUID.ValueString(),
	}

	err := r.client.CreateManagedObjectStorageCustomDomain(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageStaticSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data staticSiteModel
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
	customDomain, err := r.client.GetManagedObjectStorageCustomDomain(ctx, &request.GetManagedObjectStorageCustomDomainRequest{
		DomainName:  domainName,
		ServiceUUID: serviceUUID,
	})
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

	data.DomainName = types.StringValue(customDomain.DomainName)
	data.Type = types.StringValue(customDomain.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageStaticSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state staticSiteModel
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

	customDomain, err := r.client.ModifyManagedObjectStorageCustomDomain(ctx, &request.ModifyManagedObjectStorageCustomDomainRequest{
		DomainName:  domainName,
		ServiceUUID: serviceUUID,
		CustomDomain: request.ModifyCustomDomain{
			DomainName: data.DomainName.ValueString(),
			Type:       data.Type.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))
	data.DomainName = types.StringValue(customDomain.DomainName)
	data.Type = types.StringValue(customDomain.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageStaticSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data staticSiteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var serviceUUID, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(data.ID.ValueString(), &serviceUUID, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedObjectStorageCustomDomain(ctx, &request.DeleteManagedObjectStorageCustomDomainRequest{
		ServiceUUID: serviceUUID,
		DomainName:  domainName,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage custom domain",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *managedObjectStorageStaticSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
*/
