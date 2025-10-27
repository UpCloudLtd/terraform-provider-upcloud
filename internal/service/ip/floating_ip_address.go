package ip

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	_ resource.Resource                = &floatingIPResource{}
	_ resource.ResourceWithConfigure   = &floatingIPResource{}
	_ resource.ResourceWithImportState = &floatingIPResource{}
)

func NewFloatingIPAddressResource() resource.Resource {
	return &floatingIPResource{}
}

type floatingIPResource struct {
	client *service.Service
}

func (r *floatingIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip_address"
}

// Configure adds the provider configured client to the resource.
func (r *floatingIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type floatingIPModel struct {
	ID            types.String `tfsdk:"id"`
	Address       types.String `tfsdk:"ip_address"`
	Access        types.String `tfsdk:"access"`
	Family        types.String `tfsdk:"family"`
	MAC           types.String `tfsdk:"mac_address"`
	ReleasePolicy types.String `tfsdk:"release_policy"`
	Zone          types.String `tfsdk:"zone"`
}

func (r *floatingIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents a UpCloud floating IP address resource.",
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "An UpCloud assigned IP Address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the floating IP address. Contains the same value as `ip_address`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access": schema.StringAttribute{
				MarkdownDescription: "Network access for the floating IP address. Supported value: `public`.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("public"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"public",
					),
				},
			},
			"family": schema.StringAttribute{
				MarkdownDescription: "The address family of the floating IP address.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("IPv4"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"IPv4",
						"IPv6",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "MAC address of a server interface to assign address to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					validatorutil.NewFrameworkStringValidator(validation.IsMACAddress),
				},
			},
			"release_policy": schema.StringAttribute{
				MarkdownDescription: "The release policy of the floating IP address.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.IPAddressReleasePolicyKeep),
						string(upcloud.IPAddressReleasePolicyRelease),
					),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone of the address, e.g. `de-fra1`. Required when assigning a detached floating IP address. You can list available zones with `upctl zone list`.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func setValues(data *floatingIPModel, ip *upcloud.IPAddress) {
	data.ID = types.StringValue(ip.Address)
	data.Address = types.StringValue(ip.Address)
	data.Access = types.StringValue(ip.Access)
	data.Family = types.StringValue(ip.Family)
	data.ReleasePolicy = types.StringValue(string(ip.ReleasePolicy))
	data.Zone = types.StringValue(ip.Zone)

	if ip.MAC == "" {
		data.MAC = types.StringNull()
	} else {
		data.MAC = types.StringValue(ip.MAC)
	}
}

func (r *floatingIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data floatingIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.AssignIPAddressRequest{
		Floating:      upcloud.True,
		Access:        data.Access.ValueString(),
		Family:        data.Family.ValueString(),
		MAC:           data.MAC.ValueString(),
		ReleasePolicy: upcloud.IPAddressReleasePolicy(data.ReleasePolicy.ValueString()),
		Zone:          data.Zone.ValueString(),
	}

	ip, err := r.client.AssignIPAddress(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create floating IP address",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	setValues(&data, ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *floatingIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data floatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	ip, err := r.client.GetIPAddressDetails(ctx, &request.GetIPAddressDetailsRequest{
		Address: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read floating IP address details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	setValues(&data, ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *floatingIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state floatingIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.ModifyIPAddressRequest{
		IPAddress:     data.ID.ValueString(),
		ReleasePolicy: upcloud.IPAddressReleasePolicy(data.ReleasePolicy.ValueString()),
	}

	if !data.MAC.Equal(state.MAC) {
		apiReq.MAC = data.MAC.ValueString()
	}

	ip1, err := r.client.ModifyIPAddress(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify floating IP address",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	setValues(&data, ip1)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *floatingIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data floatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if data.MAC.ValueString() != "" {
		_, err := r.client.ModifyIPAddress(ctx, &request.ModifyIPAddressRequest{
			IPAddress: data.ID.ValueString(),
			MAC:       "",
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to detach floating IP address from a server network interface",
				utils.ErrorDiagnosticDetail(err),
			)
		}
	}

	err := r.client.ReleaseIPAddress(ctx, &request.ReleaseIPAddressRequest{
		IPAddress: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete floating IP address",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *floatingIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
