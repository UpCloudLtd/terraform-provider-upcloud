package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type backendMemberResource struct {
	client *service.Service
}

type backendMemberModel struct {
	Backend     types.String `tfsdk:"backend"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ID          types.String `tfsdk:"id"`
	IP          types.String `tfsdk:"ip"`
	MaxSessions types.Int64  `tfsdk:"max_sessions"`
	Name        types.String `tfsdk:"name"`
	Port        types.Int64  `tfsdk:"port"`
	Weight      types.Int64  `tfsdk:"weight"`
}

func backendMemberSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backend": schema.StringAttribute{
				MarkdownDescription: "ID of the load balancer backend to which the member is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the member is enabled. Disabled members are excluded from load balancing.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the backend member. ID is in `{load balancer UUID}/{backend name}/{member name}` format.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "Optional fallback IP address in case of failure on DNS resolving.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					validatorutil.NewFrameworkStringValidator(validation.IsIPAddress),
				},
			},
			"max_sessions": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of sessions before queueing.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 500000),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the member. Must be unique within within the load balancer backend.",
				Required:            true,
				Validators: []validator.String{
					validatorutil.IsDomainName(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Server port. Port is optional and can be specified in DNS SRV record.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					validatorutil.NewFrameworkInt64Validator(validation.IsPortNumber),
				},
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Weight of the member. The higher the weight, the more traffic the member receives.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
		},
	}
}

func setBackendMemberValues(_ context.Context, data *backendMemberModel, member *upcloud.LoadBalancerBackendMember) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	var loadBalancer, backend, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &backend, &name)
	if err != nil {
		diags.AddError(
			"Unable to unmarshal loadbalancer backend member ID",
			utils.ErrorDiagnosticDetail(err),
		)
		respDiagnostics.Append(diags...)
	}

	data.Backend = types.StringValue(utils.MarshalID(loadBalancer, backend))
	data.Enabled = types.BoolValue(member.Enabled)
	data.IP = types.StringValue(member.IP)
	data.MaxSessions = types.Int64Value(int64(member.MaxSessions))
	data.Name = types.StringValue(name)
	data.Port = types.Int64Value(int64(member.Port))
	data.Weight = types.Int64Value(int64(member.Weight))

	return respDiagnostics
}

func (r *backendMemberResource) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, loadBalancerType upcloud.LoadBalancerBackendMemberType) {
	var data backendMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, backendName string
	err := utils.UnmarshalID(data.Backend.ValueString(), &loadbalancer, &backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend member name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := request.CreateLoadBalancerBackendMemberRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Member: request.LoadBalancerBackendMember{
			Enabled:     data.Enabled.ValueBool(),
			IP:          data.IP.ValueString(),
			MaxSessions: int(data.MaxSessions.ValueInt64()),
			Name:        data.Name.ValueString(),
			Port:        int(data.Port.ValueInt64()),
			Type:        loadBalancerType,
			Weight:      int(data.Weight.ValueInt64()),
		},
	}

	member, err := r.client.CreateLoadBalancerBackendMember(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer backend member",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(utils.MarshalID(loadbalancer, backendName, member.Name))

	resp.Diagnostics.Append(setBackendMemberValues(ctx, &data, member)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendMemberResource) read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data backendMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadbalancer, backendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend member name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	member, err := r.client.GetLoadBalancerBackendMember(ctx, &request.GetLoadBalancerBackendMemberRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        name,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer backend member details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setBackendMemberValues(ctx, &data, member)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendMemberResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data backendMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadbalancer, backendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend member name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	apiReq := &request.ModifyLoadBalancerBackendMemberRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        name,
		Member: request.ModifyLoadBalancerBackendMember{
			Name:        name,
			Weight:      upcloud.IntPtr(int(data.Weight.ValueInt64())),
			MaxSessions: upcloud.IntPtr(int(data.MaxSessions.ValueInt64())),
			Enabled:     data.Enabled.ValueBoolPointer(),
			IP:          data.IP.ValueStringPointer(),
			Port:        int(data.Port.ValueInt64()),
		},
	}

	member, err := r.client.ModifyLoadBalancerBackendMember(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer backend member config",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(loadbalancer, backendName, member.Name))

	resp.Diagnostics.Append(setBackendMemberValues(ctx, &data, member)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *backendMemberResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data backendMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var loadbalancer, backendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &backendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer backend member name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	if err := r.client.DeleteLoadBalancerBackendMember(ctx, &request.DeleteLoadBalancerBackendMemberRequest{
		ServiceUUID: loadbalancer,
		BackendName: backendName,
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer backend member",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}
