package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &resolverResource{}
	_ resource.ResourceWithConfigure   = &resolverResource{}
	_ resource.ResourceWithImportState = &resolverResource{}
)

func NewResolverResource() resource.Resource {
	return &resolverResource{}
}

type resolverResource struct {
	client *service.Service
}

func (r *resolverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_resolver"
}

// Configure adds the provider configured client to the resource.
func (r *resolverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type resolverModel struct {
	CacheInvalid types.Int64  `tfsdk:"cache_invalid"`
	CacheValid   types.Int64  `tfsdk:"cache_valid"`
	ID           types.String `tfsdk:"id"`
	LoadBalancer types.String `tfsdk:"loadbalancer"`
	Name         types.String `tfsdk:"name"`
	Nameservers  types.List   `tfsdk:"nameservers"`
	Retries      types.Int64  `tfsdk:"retries"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	TimeoutRetry types.Int64  `tfsdk:"timeout_retry"`
}

func (r *resolverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents load balancer resolver.",
		Attributes: map[string]schema.Attribute{
			"cache_invalid": schema.Int64Attribute{
				MarkdownDescription: "Time in seconds to cache invalid results.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"cache_valid": schema.Int64Attribute{
				MarkdownDescription: "Time in seconds to cache valid results.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the resolver. ID is in `{load balancer UUID}/{resolver name}` format.",
				Computed:            true,
			},
			"loadbalancer": schema.StringAttribute{
				MarkdownDescription: "ID of the load balancer to which the resolver is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the resolver. Must be unique within the service.",
				Required:            true,
				Validators: []validator.String{
					nameValidator,
				},
			},
			"nameservers": schema.ListAttribute{
				MarkdownDescription: `List of nameserver IP addresses. Nameserver can reside in public internet or in customer private network. Port is optional, if missing then default 53 will be used.`,
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 10),
				},
			},
			"retries": schema.Int64Attribute{
				MarkdownDescription: "Number of retries on failure.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for the query in seconds.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 60),
				},
			},
			"timeout_retry": schema.Int64Attribute{
				MarkdownDescription: "Timeout for the query retries in seconds.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 60),
				},
			},
		},
	}
}

func setResolverValues(ctx context.Context, data *resolverModel, resolver *upcloud.LoadBalancerResolver) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	var loadBalancer, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &name)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to unmarshal loadbalancer resolver ID",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.LoadBalancer = types.StringValue(loadBalancer)
	data.Name = types.StringValue(name)

	data.CacheInvalid = types.Int64Value(int64(resolver.CacheInvalid))
	data.CacheValid = types.Int64Value(int64(resolver.CacheValid))

	data.Nameservers, diags = types.ListValueFrom(ctx, types.StringType, resolver.Nameservers)
	respDiagnostics.Append(diags...)

	data.Retries = types.Int64Value(int64(resolver.Retries))
	data.Timeout = types.Int64Value(int64(resolver.Timeout))
	data.TimeoutRetry = types.Int64Value(int64(resolver.TimeoutRetry))

	return respDiagnostics
}

func (r *resolverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resolverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var nameservers []string
	if !data.Nameservers.IsNull() && !data.Nameservers.IsUnknown() {
		resp.Diagnostics.Append(data.Nameservers.ElementsAs(ctx, &nameservers, false)...)
	}

	apiReq := request.CreateLoadBalancerResolverRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Resolver: request.LoadBalancerResolver{
			Name:         data.Name.ValueString(),
			Nameservers:  nameservers,
			Retries:      int(data.Retries.ValueInt64()),
			Timeout:      int(data.Timeout.ValueInt64()),
			TimeoutRetry: int(data.TimeoutRetry.ValueInt64()),
			CacheValid:   int(data.CacheValid.ValueInt64()),
			CacheInvalid: int(data.CacheInvalid.ValueInt64()),
		},
	}

	resolver, err := r.client.CreateLoadBalancerResolver(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer resolver",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.LoadBalancer.ValueString(), data.Name.ValueString()))

	resp.Diagnostics.Append(setResolverValues(ctx, &data, resolver)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resolverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resolverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadbalancer, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer resolver ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resolver, err := r.client.GetLoadBalancerResolver(ctx, &request.GetLoadBalancerResolverRequest{
		Name:        name,
		ServiceUUID: loadbalancer,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer resolver details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setResolverValues(ctx, &data, resolver)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resolverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resolverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var id string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	var loadbalancer, name string
	err := utils.UnmarshalID(id, &loadbalancer, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer resolver ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	var nameservers []string
	if !data.Nameservers.IsNull() && !data.Nameservers.IsUnknown() {
		resp.Diagnostics.Append(data.Nameservers.ElementsAs(ctx, &nameservers, false)...)
	}

	apiReq := request.ModifyLoadBalancerResolverRequest{
		ServiceUUID: loadbalancer,
		Name:        name,
		Resolver: request.LoadBalancerResolver{
			Name:         data.Name.ValueString(),
			Nameservers:  nameservers,
			Retries:      int(data.Retries.ValueInt64()),
			Timeout:      int(data.Timeout.ValueInt64()),
			TimeoutRetry: int(data.TimeoutRetry.ValueInt64()),
			CacheValid:   int(data.CacheValid.ValueInt64()),
			CacheInvalid: int(data.CacheInvalid.ValueInt64()),
		},
	}

	resolver, err := r.client.ModifyLoadBalancerResolver(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer resolver",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(loadbalancer, resolver.Name))

	resp.Diagnostics.Append(setResolverValues(ctx, &data, resolver)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resolverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resolverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoadBalancerResolver(ctx, &request.DeleteLoadBalancerResolverRequest{
		ServiceUUID: data.LoadBalancer.ValueString(),
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer resolver",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *resolverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
