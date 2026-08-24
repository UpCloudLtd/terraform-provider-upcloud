package tag

import (
	"context"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
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

const (
	nameDescription        = "The name of the tag."
	descriptionDescription = "Free form text representing the meaning of the tag."
	serversDescription     = "A collection of servers that have been assigned the tag."
)

var (
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *service.Service
}

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

// Configure adds the provider configured client to the resource.
func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type tagCommonModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Servers     types.Set    `tfsdk:"servers"`
}

type tagModel struct {
	tagCommonModel

	ID types.String `tfsdk:"id"`
}

func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `~> Consider using labels instead of tags. Tags are an access control feature and only available for a limited set of resources. Use labels to describe and filter your resources.

Resource for managing tags. When tagging multiple servers with the same tag, use this resource to create the tag and ` + "`" + `tags` + "`" + ` field of the server resource to tag the server.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tag. Contains the same value as `name`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[a-zA-Z0-9-]+$"),
						"must contain only alphanumeric characters and hyphens",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription,
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 255),
				},
			},
			"servers": schema.SetAttribute{
				MarkdownDescription: serversDescription,
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func setValues(ctx context.Context, data *tagCommonModel, tag *upcloud.Tag) diag.Diagnostics {
	data.Name = types.StringValue(tag.Name)
	data.Description = types.StringValue(tag.Description)

	var diags diag.Diagnostics
	if data.Servers.IsNull() {
		data.Servers = types.SetNull(types.StringType)
	} else {
		data.Servers, diags = types.SetValueFrom(ctx, types.StringType, utils.NilAsEmptyList(tag.Servers))
	}
	return diags
}

func commonModelToTag(ctx context.Context, data tagCommonModel) (upcloud.Tag, diag.Diagnostics) {
	var servers []string
	if !data.Servers.IsNull() {
		diags := data.Servers.ElementsAs(ctx, &servers, false)
		if diags.HasError() {
			return upcloud.Tag{}, diags
		}
	}

	return upcloud.Tag{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Servers:     servers,
	}, nil
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	reqTag, diags := commonModelToTag(ctx, data.tagCommonModel)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var servers []string
	if !data.Servers.IsNull() {
		resp.Diagnostics.Append(data.Servers.ElementsAs(ctx, &servers, false)...)
	}

	apiReq := request.CreateTagRequest{
		Tag: reqTag,
	}

	tag, err := r.client.CreateTag(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create tag",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(tag.Name)

	resp.Diagnostics.Append(setValues(ctx, &data.tagCommonModel, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getTagByName(ctx context.Context, client *service.Service, name string) (*upcloud.Tag, diag.Diagnostics) {
	var diags diag.Diagnostics

	tags, err := client.GetTags(ctx)
	if err != nil {
		diags.AddError(
			"Unable to read tags",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	for _, tag := range tags.Tags {
		if tag.Name == name {
			return &tag, diags
		}
	}

	return nil, diags
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data tagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	tag, diags := getTagByName(ctx, r.client, data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if tag == nil && !diags.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data.tagCommonModel, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data tagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	reqTag, diags := commonModelToTag(ctx, data.tagCommonModel)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if reqTag.Servers == nil {
		// Tags are modified with PUT request. Get current value for servers field to avoid untagging servers when modifying description.
		tag, diags := getTagByName(ctx, r.client, data.ID.ValueString())
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		reqTag.Servers = tag.Servers
	}

	apiReq := request.ModifyTagRequest{
		Name: data.ID.ValueString(),
		Tag:  reqTag,
	}

	tag, err := r.client.ModifyTag(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update tag",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data.tagCommonModel, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteTag(ctx, &request.DeleteTagRequest{
		Name: data.ID.ValueString(),
	})
	if err != nil && !utils.IsNotFoundError(err) {
		resp.Diagnostics.AddError(
			"Unable to delete tag",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
