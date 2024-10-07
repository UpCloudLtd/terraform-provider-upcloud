package loadbalancer

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &frontendRuleResource{}
	_ resource.ResourceWithConfigure   = &frontendRuleResource{}
	_ resource.ResourceWithImportState = &frontendRuleResource{}
)

func NewFrontendRuleResource() resource.Resource {
	return &frontendRuleResource{}
}

type frontendRuleResource struct {
	client *service.Service
}

func (r *frontendRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_frontend_rule"
}

// Configure adds the provider configured client to the resource.
func (r *frontendRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type frontendRuleModel struct {
	ID       types.String `tfsdk:"id"`
	Frontend types.String `tfsdk:"frontend"`
	Name     types.String `tfsdk:"name"`
	Priority types.Int64  `tfsdk:"priority"`
	Matchers types.List   `tfsdk:"matchers"`
	Actions  types.List   `tfsdk:"actions"`
}

func (r *frontendRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents load balancer frontend rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the frontend rule. ID is in `{load balancer UUID}/{frontend name}/{frontend rule name}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"frontend": schema.StringAttribute{
				MarkdownDescription: "ID of the load balancer frontend to which the frontend rule is connected.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the frontend rule. Must be unique within the frontend.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					nameValidator,
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Rule with the higher priority goes first. Rules with the same priority processed in alphabetical order.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"actions": schema.ListNestedBlock{
				MarkdownDescription: "Rule actions.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"http_redirect": schema.ListNestedBlock{
							MarkdownDescription: "Redirects HTTP requests to specified location or URL scheme. Only either location or scheme can be defined at a time.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"location": schema.StringAttribute{
										MarkdownDescription: "Target location.",
										Optional:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("scheme")),
										},
									},
									"scheme": schema.StringAttribute{
										MarkdownDescription: "Target scheme.",
										Optional:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.OneOf(
												string(upcloud.LoadBalancerActionHTTPRedirectSchemeHTTP),
												string(upcloud.LoadBalancerActionHTTPRedirectSchemeHTTPS),
											),
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("location")),
										},
									},
								},
							},
						},
						"http_return": schema.ListNestedBlock{
							MarkdownDescription: "Returns HTTP response with specified HTTP status.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										MarkdownDescription: "Content type.",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"status": schema.Int64Attribute{
										MarkdownDescription: "HTTP status code.",
										Required:            true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
										Validators: []validator.Int64{
											int64validator.Between(100, 599),
										},
									},
									"payload": schema.StringAttribute{
										MarkdownDescription: "The payload.",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 4096),
										},
									},
								},
							},
						},
						"set_forwarded_headers": schema.ListNestedBlock{
							MarkdownDescription: "Adds 'X-Forwarded-For / -Proto / -Port' headers in your forwarded requests",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"active": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										Default:  booldefault.StaticBool(true),
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"tcp_reject": schema.ListNestedBlock{
							MarkdownDescription: "Terminates a connection.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"active": schema.BoolAttribute{
										MarkdownDescription: "Indicates if the rule is active.",
										Optional:            true,
										Computed:            true,
										Default:             booldefault.StaticBool(true),
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"use_backend": schema.ListNestedBlock{
							MarkdownDescription: "Routes traffic to specified `backend`.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"backend_name": schema.StringAttribute{
										MarkdownDescription: "The name of the backend where traffic will be routed.",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
			},
			"matchers": schema.ListNestedBlock{
				MarkdownDescription: "Set of rule matchers. If rule doesn't have matchers, then action applies to all incoming requests.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"body_size": schema.ListNestedBlock{
							MarkdownDescription: "Matches by HTTP request body size.",
							NestedObject:        frontendRuleMatcherIntegerSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"body_size_range": schema.ListNestedBlock{
							MarkdownDescription: "Matches by range of HTTP request body sizes.",
							NestedObject:        frontendRuleMatcherRangeSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"cookie": schema.ListNestedBlock{
							MarkdownDescription: "Matches by HTTP cookie value. Cookie name must be provided.",
							NestedObject:        frontendRuleMatcherStringWithArgumentSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"header": schema.ListNestedBlock{
							MarkdownDescription: "Matches by HTTP header value. Header name must be provided.",
							NestedObject:        frontendRuleMatcherStringWithArgumentSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"host": schema.ListNestedBlock{
							MarkdownDescription: "Matches by hostname. Header extracted from HTTP Headers or from TLS certificate in case of secured connection.",
							NestedObject:        frontendRuleMatcherHostSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"http_method": schema.ListNestedBlock{
							MarkdownDescription: "Matches by HTTP method.",
							NestedObject:        frontendRuleMatcherHTTPMethodSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"num_members_up": schema.ListNestedBlock{
							MarkdownDescription: "Matches by number of healthy backend members.",
							NestedObject:        frontendRuleMatcherBackendSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"path": schema.ListNestedBlock{
							MarkdownDescription: "Matches by URL path.",
							NestedObject:        frontendRuleMatcherStringSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"src_ip": schema.ListNestedBlock{
							MarkdownDescription: "Matches by source IP address.",
							NestedObject:        frontendRuleMatcherIPSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"src_port": schema.ListNestedBlock{
							MarkdownDescription: "Matches by source port number.",
							NestedObject:        frontendRuleMatcherIntegerSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"src_port_range": schema.ListNestedBlock{
							MarkdownDescription: "Matches by range of source port numbers.",
							NestedObject:        frontendRuleMatcherRangeSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"url": schema.ListNestedBlock{
							MarkdownDescription: "Matches by URL without schema, e.g. `example.com/dashboard`.",
							NestedObject:        frontendRuleMatcherStringSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"url_param": schema.ListNestedBlock{
							MarkdownDescription: "Matches by URL query parameter value. Query parameter name must be provided",
							NestedObject:        frontendRuleMatcherStringWithArgumentSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
						"url_query": schema.ListNestedBlock{
							MarkdownDescription: "Matches by URL query string.",
							NestedObject:        frontendRuleMatcherStringSchema(),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 100),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
			},
		},
	}
}

func setFrontendRuleValues(ctx context.Context, data *frontendRuleModel, frontendRule *upcloud.LoadBalancerFrontendRule, blocks map[string]schema.ListNestedBlock) diag.Diagnostics {
	var respDiagnostics diag.Diagnostics

	isImport := data.Frontend.ValueString() == ""

	var loadBalancer, frontendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &frontendName, &name)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend rule ID",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.Frontend = types.StringValue(utils.MarshalID(loadBalancer, frontendName))
	data.Name = types.StringValue(name)
	data.Priority = types.Int64Value(int64(frontendRule.Priority))

	if !data.Actions.IsNull() || isImport {
		respDiagnostics.Append(setFrontendRuleActionsValues(ctx, data, frontendRule, blocks)...)
	}

	if !data.Matchers.IsNull() || isImport {
		respDiagnostics.Append(setFrontendRuleMatchersValues(ctx, data, frontendRule, blocks)...)
	}

	return respDiagnostics
}

func (r *frontendRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data frontendRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var loadbalancer, frontendName string
	err := utils.UnmarshalID(data.Frontend.ValueString(), &loadbalancer, &frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	matchers, diags := buildFrontendRuleMatchers(ctx, data.Matchers)
	resp.Diagnostics.Append(diags...)

	actions, diags := buildFrontendRuleActions(ctx, data.Actions)
	resp.Diagnostics.Append(diags...)

	apiReq := request.CreateLoadBalancerFrontendRuleRequest{
		ServiceUUID:  loadbalancer,
		FrontendName: frontendName,
		Rule: request.LoadBalancerFrontendRule{
			Name:     data.Name.ValueString(),
			Priority: int(data.Priority.ValueInt64()),
			Matchers: matchers,
			Actions:  actions,
		},
	}

	frontendRule, err := r.client.CreateLoadBalancerFrontendRule(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create loadbalancer frontend rule",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.Frontend.ValueString(), data.Name.ValueString()))

	blocks := make(map[string]schema.ListNestedBlock)
	for k, v := range req.Config.Schema.GetBlocks() {
		block, ok := v.(schema.ListNestedBlock)
		if !ok {
			continue
		}

		blocks[k] = block
	}

	resp.Diagnostics.Append(setFrontendRuleValues(ctx, &data, frontendRule, blocks)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data frontendRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var loadbalancer, frontendName, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &loadbalancer, &frontendName, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend rule ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	frontendRule, err := r.client.GetLoadBalancerFrontendRule(ctx, &request.GetLoadBalancerFrontendRuleRequest{
		FrontendName: frontendName,
		Name:         name,
		ServiceUUID:  loadbalancer,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read loadbalancer frontend rule details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	blocks := make(map[string]schema.ListNestedBlock)
	for k, v := range req.State.Schema.GetBlocks() {
		block, ok := v.(schema.ListNestedBlock)
		if !ok {
			continue
		}

		blocks[k] = block
	}

	resp.Diagnostics.Append(setFrontendRuleValues(ctx, &data, frontendRule, blocks)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data frontendRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var loadBalancer, frontendName, name string
	if err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &frontendName, &name); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend rule ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.ModifyLoadBalancerFrontendRuleRequest{
		ServiceUUID:  loadBalancer,
		FrontendName: frontendName,
		Name:         name,
		Rule: request.ModifyLoadBalancerFrontendRule{
			Name:     data.Name.ValueString(),
			Priority: upcloud.IntPtr(int(data.Priority.ValueInt64())),
		},
	}

	frontendRule, err := r.client.ModifyLoadBalancerFrontendRule(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify loadbalancer frontend rule",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	blocks := make(map[string]schema.ListNestedBlock)
	for k, v := range req.Config.Schema.GetBlocks() {
		block, ok := v.(schema.ListNestedBlock)
		if !ok {
			continue
		}

		blocks[k] = block
	}

	resp.Diagnostics.Append(setFrontendRuleValues(ctx, &data, frontendRule, blocks)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *frontendRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data frontendRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var loadBalancer, frontendName, name string
	if err := utils.UnmarshalID(data.ID.ValueString(), &loadBalancer, &frontendName, &name); err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal loadbalancer frontend rule ID",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoadBalancerFrontendRule(ctx, &request.DeleteLoadBalancerFrontendRuleRequest{
		ServiceUUID:  loadBalancer,
		FrontendName: frontendName,
		Name:         name,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete loadbalancer frontend rule",
			utils.ErrorDiagnosticDetail(err),
		)
	}
}

func (r *frontendRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func elementTypesByKey(k string, blocks map[string]schema.ListNestedBlock) (map[string]basetypes.ObjectTypable, error) {
	topLevel, ok := blocks[k]
	if !ok {
		return nil, fmt.Errorf("block by key %s not found", k)
	}

	elementTypes := make(map[string]basetypes.ObjectTypable)

	for blockName, b := range topLevel.NestedObject.Blocks {
		block, ok := b.(schema.ListNestedBlock)
		if !ok {
			continue
		}

		elementTypes[blockName] = block.NestedObject.Type()
	}

	return elementTypes, nil
}
