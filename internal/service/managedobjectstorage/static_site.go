package managedobjectstorage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &managedObjectStorageStaticSiteResource{}
	_ resource.ResourceWithConfigure   = &managedObjectStorageStaticSiteResource{}
	_ resource.ResourceWithImportState = &managedObjectStorageStaticSiteResource{}
)

func NewStaticSiteResource() resource.Resource {
	return &managedObjectStorageStaticSiteResource{}
}

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
	ErrorPages    types.List   `tfsdk:"error_page"`
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
					errorPagesValidator{},
					listvalidator.SizeBetween(0, 25),
				},
			},
		},
	}
}

func errorPageType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"status_code":        types.Int64Type,
		"status_range_start": types.Int64Type,
		"status_range_end":   types.Int64Type,
		"error_document":     types.StringType,
	}}
}

func parseErrorPages(ctx context.Context, list types.List) ([]v9.ObjectStorage2StaticWebsiteErrorPage, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var pages []errorPageModel
	if diags := list.ElementsAs(ctx, &pages, false); diags.HasError() {
		return nil, fmt.Errorf("failed to decode error_page block")
	}

	if err := validateErrorPageStatusMatcherAtIndex(pages); err != nil {
		return nil, err
	}

	result := make([]v9.ObjectStorage2StaticWebsiteErrorPage, 0, len(pages))
	for _, page := range pages {
		p := v9.ObjectStorage2StaticWebsiteErrorPage{
			ErrorDocument: page.ErrorDocument.ValueString(),
		}

		if !page.StatusCode.IsNull() && !page.StatusCode.IsUnknown() {
			v := int(page.StatusCode.ValueInt64())
			p.StatusCode = &v
		}

		if !page.StatusRangeStart.IsNull() && !page.StatusRangeStart.IsUnknown() && !page.StatusRangeEnd.IsNull() && !page.StatusRangeEnd.IsUnknown() {
			p.StatusRange = &struct {
				End   int `json:"end"`
				Start int `json:"start"`
			}{
				Start: int(page.StatusRangeStart.ValueInt64()),
				End:   int(page.StatusRangeEnd.ValueInt64()),
			}
		}

		result = append(result, p)
	}

	return result, nil
}

func setStaticSiteValues(ctx context.Context, data *staticSiteModel, site *v9.ObjectStorage2StaticWebsiteConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	data.DomainName = types.StringValue(site.DomainName)
	data.BucketName = types.StringValue(site.BucketName)
	data.BucketPrefix = types.StringValue(site.BucketPrefix)
	data.IndexDocument = types.StringValue(site.IndexDocument)
	data.Enabled = types.BoolValue(site.Enabled)
	data.SpaMode = types.BoolPointerValue(site.SpaMode)

	pages := make([]errorPageModel, 0, len(site.ErrorPages))
	for _, page := range site.ErrorPages {
		item := errorPageModel{
			ErrorDocument:    types.StringValue(page.ErrorDocument),
			StatusCode:       types.Int64Null(),
			StatusRangeStart: types.Int64Null(),
			StatusRangeEnd:   types.Int64Null(),
		}

		if page.StatusCode != nil {
			item.StatusCode = types.Int64Value(int64(*page.StatusCode))
		}

		if page.StatusRange != nil {
			item.StatusRangeStart = types.Int64Value(int64(page.StatusRange.Start))
			item.StatusRangeEnd = types.Int64Value(int64(page.StatusRange.End))
		}

		pages = append(pages, item)
	}

	errorPages, pageDiags := types.ListValueFrom(ctx, errorPageType(), pages)
	diags.Append(pageDiags...)
	data.ErrorPages = errorPages

	return diags
}

func parseServiceUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func diagUnexpectedStatus(respDiags *diag.Diagnostics, action string, status int, body []byte) {
	respDiags.AddError(
		fmt.Sprintf("Unable to %s managed object storage static site", action),
		fmt.Sprintf("Unexpected API status code %d: %s", status, string(body)),
	)
}

func (r *managedObjectStorageStaticSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data staticSiteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateErrorPages(ctx, data.ErrorPages, path.Root("error_page"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := parseServiceUUID(data.ServiceUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	errorPages, err := parseErrorPages(ctx, data.ErrorPages)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse static site error pages",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	body := v9.CreateObjectStorageStaticWebsiteJSONRequestBody{
		BucketName: data.BucketName.ValueString(),
	}

	if !data.DomainName.IsNull() && !data.DomainName.IsUnknown() && data.DomainName.ValueString() != "" {
		body.DomainName = data.DomainName.ValueStringPointer()
	}
	if !data.BucketPrefix.IsNull() && !data.BucketPrefix.IsUnknown() {
		body.BucketPrefix = data.BucketPrefix.ValueStringPointer()
	}
	if !data.IndexDocument.IsNull() && !data.IndexDocument.IsUnknown() {
		body.IndexDocument = data.IndexDocument.ValueStringPointer()
	}
	if !data.SpaMode.IsNull() && !data.SpaMode.IsUnknown() {
		body.SpaMode = data.SpaMode.ValueBoolPointer()
	}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		body.Enabled = data.Enabled.ValueBoolPointer()
	}
	if errorPages != nil {
		body.ErrorPages = &errorPages
	}

	created, err := r.client.CreateObjectStorageStaticWebsiteWithResponse(ctx, svcUUID, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage static site",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}
	if created.StatusCode() != http.StatusCreated {
		diagUnexpectedStatus(&resp.Diagnostics, "create", created.StatusCode(), created.Body)
		return
	}

	if created.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Unable to create managed object storage static site",
			utils.ErrorDiagnosticDetail(fmt.Errorf("unexpected response: %s", created.HTTPResponse.Status)),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), created.JSON201.DomainName))
	resp.Diagnostics.Append(setStaticSiteValues(ctx, &data, created.JSON201)...)

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
	svcUUID, err := parseServiceUUID(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	site, err := r.client.GetObjectStorageStaticWebsiteWithResponse(ctx, svcUUID, domainName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed object storage static site details",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if site.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if site.StatusCode() != http.StatusOK || site.JSON200 == nil {
		diagUnexpectedStatus(&resp.Diagnostics, "read", site.StatusCode(), site.Body)
		return
	}

	resp.Diagnostics.Append(setStaticSiteValues(ctx, &data, site.JSON200)...)
	data.ID = types.StringValue(utils.MarshalID(serviceUUID, data.DomainName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedObjectStorageStaticSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state staticSiteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateErrorPages(ctx, data.ErrorPages, path.Root("error_page"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUUID, domainName string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &serviceUUID, &domainName)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcUUID, err := parseServiceUUID(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	errorPages, err := parseErrorPages(ctx, data.ErrorPages)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse static site error pages",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	body := v9.ModifyObjectStorageStaticWebsiteJSONRequestBody{}
	if !data.BucketName.IsNull() && !data.BucketName.IsUnknown() {
		body.BucketName = data.BucketName.ValueStringPointer()
	}
	if !data.BucketPrefix.IsNull() && !data.BucketPrefix.IsUnknown() {
		body.BucketPrefix = data.BucketPrefix.ValueStringPointer()
	}
	if !data.IndexDocument.IsNull() && !data.IndexDocument.IsUnknown() {
		body.IndexDocument = data.IndexDocument.ValueStringPointer()
	}
	if !data.SpaMode.IsNull() && !data.SpaMode.IsUnknown() {
		body.SpaMode = data.SpaMode.ValueBoolPointer()
	}
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		body.Enabled = data.Enabled.ValueBoolPointer()
	}
	if errorPages != nil {
		body.ErrorPages = &errorPages
	}

	updated, err := r.client.ModifyObjectStorageStaticWebsiteWithResponse(ctx, svcUUID, domainName, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify managed object storage static site",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if updated.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if updated.StatusCode() != http.StatusOK || updated.JSON200 == nil {
		diagUnexpectedStatus(&resp.Diagnostics, "modify", updated.StatusCode(), updated.Body)
		return
	}

	resp.Diagnostics.Append(setStaticSiteValues(ctx, &data, updated.JSON200)...)
	data.ID = types.StringValue(utils.MarshalID(data.ServiceUUID.ValueString(), data.DomainName.ValueString()))
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

	svcUUID, err := parseServiceUUID(serviceUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse service UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	deleted, err := r.client.DeleteObjectStorageStaticWebsiteWithResponse(ctx, svcUUID, domainName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed object storage static site",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if deleted.StatusCode() != http.StatusNoContent && deleted.StatusCode() != http.StatusNotFound {
		diagUnexpectedStatus(&resp.Diagnostics, "delete", deleted.StatusCode(), deleted.Body)
	}
}

func (r *managedObjectStorageStaticSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
