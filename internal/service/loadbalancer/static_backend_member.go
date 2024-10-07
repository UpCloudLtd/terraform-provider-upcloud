package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &backendStaticMemberResource{}
	_ resource.ResourceWithConfigure   = &backendStaticMemberResource{}
	_ resource.ResourceWithImportState = &backendStaticMemberResource{}
)

func NewBackendStaticMemberResource() resource.Resource {
	return &backendStaticMemberResource{}
}

type backendStaticMemberResource struct {
	backendMemberResource
}

func (r *backendStaticMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_static_backend_member"
}

// Configure adds the provider configured client to the resource.
func (r *backendStaticMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

func (r *backendStaticMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = backendMemberSchema()
	resp.Schema.MarkdownDescription = "This resource represents load balancer static backend member"
}

func (r *backendStaticMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp, upcloud.LoadBalancerBackendMemberTypeStatic)
}

func (r *backendStaticMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.read(ctx, req, resp)
}

func (r *backendStaticMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *backendStaticMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

func (r *backendStaticMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
