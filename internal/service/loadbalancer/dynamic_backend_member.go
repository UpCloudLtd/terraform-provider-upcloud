package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &backendDynamicMemberResource{}
	_ resource.ResourceWithConfigure   = &backendDynamicMemberResource{}
	_ resource.ResourceWithImportState = &backendDynamicMemberResource{}
)

func NewBackendDynamicMemberResource() resource.Resource {
	return &backendDynamicMemberResource{}
}

type backendDynamicMemberResource struct {
	backendMemberResource
}

func (r *backendDynamicMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_dynamic_backend_member"
}

// Configure adds the provider configured client to the resource.
func (r *backendDynamicMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

func (r *backendDynamicMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = backendMemberSchema()
	resp.Schema.MarkdownDescription = "This resource represents load balancer dynamic backend member"
}

func (r *backendDynamicMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp, upcloud.LoadBalancerBackendMemberTypeDynamic)
}

func (r *backendDynamicMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.read(ctx, req, resp)
}

func (r *backendDynamicMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *backendDynamicMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

func (r *backendDynamicMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
