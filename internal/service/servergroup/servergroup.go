package servergroup

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	titleDescription        = "Title of your server group"
	membersDescription      = "UUIDs of the servers that are members of this group. Servers can also be attached to the server group via `server_group` property of `upcloud_server`. See also `track_members` property."
	trackMembersDescription = "Controls if members of the server group are being tracked in this resource. Set to `false` when using `server_group` property of `upcloud_server` to attach servers to the server group to avoid delayed state updates."
	// Lines > 1 should have one level of indentation to keep them under the right list item
	antiAffinityPolicyDescription = `Defines if a server group is an anti-affinity group. Setting this to ` + "`strict` or `yes`" + ` will
	result in all servers in the group being placed on separate compute hosts. The value can be ` + "`strict`, `yes`, or `no`" + `.

	* ` + "`strict`" + ` policy doesn't allow servers in the same server group to be on the same host
	* ` + "`yes`" + ` refers to best-effort policy and tries to put servers on different hosts, but this is not guaranteed
	* ` + "`no`" + ` refers to having no policy and thus no effect on server host affinity

	To verify if the anti-affinity policies are met by requesting a server group details from API. For more information
	please see UpCloud API documentation on server groups.

	Plese also note that anti-affinity policies are only applied on server start. This means that if anti-affinity
	policies in server group are not met, you need to manually restart the servers in said group,
	for example via API, UpCloud Control Panel or upctl (UpCloud CLI)`
)

var (
	_ resource.Resource                = &serverGroupResource{}
	_ resource.ResourceWithConfigure   = &serverGroupResource{}
	_ resource.ResourceWithImportState = &serverGroupResource{}
)

func NewServerGroupResource() resource.Resource {
	return &serverGroupResource{}
}

type serverGroupResource struct {
	client *service.Service
}

func (r *serverGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_group"
}

// Configure adds the provider configured client to the resource.
func (r *serverGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type serverGroupModel struct {
	ID                 types.String `tfsdk:"id"`
	Title              types.String `tfsdk:"title"`
	Labels             types.Map    `tfsdk:"labels"`
	Members            types.Set    `tfsdk:"members"`
	AntiAffinityPolicy types.String `tfsdk:"anti_affinity_policy"`
	TrackMembers       types.Bool   `tfsdk:"track_members"`
}

func (r *serverGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Server groups allow grouping servers and defining anti-affinity for the servers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the server group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: titleDescription,
				Required:            true,
			},
			"labels": utils.LabelsAttribute("server group"),
			"members": schema.SetAttribute{
				MarkdownDescription: membersDescription,
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"track_members": schema.BoolAttribute{
				MarkdownDescription: trackMembersDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				Validators: []validator.Bool{
					trackMembersValidator{},
				},
			},
			"anti_affinity_policy": schema.StringAttribute{
				MarkdownDescription: antiAffinityPolicyDescription,
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(string(upcloud.ServerGroupAntiAffinityPolicyOff)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.ServerGroupAntiAffinityPolicyBestEffort),
						string(upcloud.ServerGroupAntiAffinityPolicyOff),
						string(upcloud.ServerGroupAntiAffinityPolicyStrict),
					),
				},
			},
		},
	}
}

func setValues(ctx context.Context, data *serverGroupModel, serverGroup *upcloud.ServerGroup) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	data.Title = types.StringValue(serverGroup.Title)
	data.ID = types.StringValue(serverGroup.UUID)
	data.AntiAffinityPolicy = types.StringValue(string(serverGroup.AntiAffinityPolicy))

	data.Labels, respDiagnostics = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(serverGroup.Labels))

	if data.TrackMembers.ValueBool() && !data.Members.IsNull() {
		var diags diag.Diagnostics
		data.Members, diags = types.SetValueFrom(ctx, types.StringType, serverGroup.Members)
		respDiagnostics.Append(diags...)
	} else {
		data.Members = types.SetNull(types.StringType)
	}

	return respDiagnostics
}

func (r *serverGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	var labelsSlice upcloud.LabelSlice = utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	var members upcloud.ServerUUIDSlice
	resp.Diagnostics.Append(data.Members.ElementsAs(ctx, &members, false)...)

	apiReq := request.CreateServerGroupRequest{
		Title:              data.Title.ValueString(),
		Labels:             &labelsSlice,
		Members:            members,
		AntiAffinityPolicy: upcloud.ServerGroupAntiAffinityPolicy(data.AntiAffinityPolicy.ValueString()),
	}

	serverGroup, err := r.client.CreateServerGroup(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create server group",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, serverGroup)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	serverGroup, err := r.client.GetServerGroup(ctx, &request.GetServerGroupRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read server group details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, serverGroup)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data serverGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	var labelsSlice upcloud.LabelSlice = utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	var members upcloud.ServerUUIDSlice
	resp.Diagnostics.Append(data.Members.ElementsAs(ctx, &members, false)...)

	apiReq := request.ModifyServerGroupRequest{
		UUID:               data.ID.ValueString(),
		Title:              data.Title.ValueString(),
		Labels:             &labelsSlice,
		Members:            &members,
		AntiAffinityPolicy: upcloud.ServerGroupAntiAffinityPolicy(data.AntiAffinityPolicy.ValueString()),
	}

	serverGroup, err := r.client.ModifyServerGroup(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify server group",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, serverGroup)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteServerGroup(ctx, &request.DeleteServerGroupRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete server group",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *serverGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
