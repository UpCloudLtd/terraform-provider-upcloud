package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type frontendRuleActionModel struct {
	HTTPRedirect        types.List `tfsdk:"http_redirect"`
	HTTPReturn          types.List `tfsdk:"http_return"`
	SetForwardedHeaders types.List `tfsdk:"set_forwarded_headers"`
	SetRequestHeader    types.List `tfsdk:"set_request_header"`
	SetResponseHeader   types.List `tfsdk:"set_response_header"`
	TCPReject           types.List `tfsdk:"tcp_reject"`
	UseBackend          types.List `tfsdk:"use_backend"`
}

type frontendRuleActionHTTPRedirectModel struct {
	Location types.String `tfsdk:"location"`
	Scheme   types.String `tfsdk:"scheme"`
	Status   types.Int64  `tfsdk:"status"`
}

type frontendRuleActionHTTPReturnModel struct {
	ContentType types.String `tfsdk:"content_type"`
	Status      types.Int64  `tfsdk:"status"`
	Payload     types.String `tfsdk:"payload"`
}

type frontendRuleActionSetForwardedHeadersModel struct {
	Active types.Bool `tfsdk:"active"`
}

type frontendRuleActionSetHeaderModel struct {
	Header types.String `tfsdk:"header"`
	Value  types.String `tfsdk:"value"`
}

type frontendRuleActionTCPRejectModel struct {
	Active types.Bool `tfsdk:"active"`
}

type frontendRuleActionUseBackendModel struct {
	BackendName types.String `tfsdk:"backend_name"`
}

func frontendRuleActionSetHeaderSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"header": schema.StringAttribute{
				MarkdownDescription: "Header name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Header value.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
				},
			},
		},
	}
}

func buildFrontendRuleActions(ctx context.Context, dataActions types.List) ([]upcloud.LoadBalancerAction, diag.Diagnostics) {
	if dataActions.IsNull() {
		return nil, nil
	}

	var planActions []frontendRuleActionModel
	diags := dataActions.ElementsAs(ctx, &planActions, false)
	if diags.HasError() {
		return nil, diags
	}

	actions := make([]upcloud.LoadBalancerAction, 0)
	for _, planAction := range planActions {
		// Ensure set_forwarded_headers action is iterated first to maintain correct action order.
		// Managed Load Balancer evaluates actions in the order they are set, but separate TF blocks can't guarantee this order.
		// This isn't a major issue since all actions except set_forwarded_headers are "final" (i.e., they end the chain).
		// The main use-case is having set_forwarded_headers first, followed by a "final" action.
		// We work around the ordering problem by always setting set_forwarded_headers actions first.
		var setForwardedHeaders []frontendRuleActionSetForwardedHeadersModel
		diags = planAction.SetForwardedHeaders.ElementsAs(ctx, &setForwardedHeaders, false)
		if diags.HasError() {
			return nil, diags
		}

		for range setForwardedHeaders {
			action := request.NewLoadBalancerSetForwardedHeadersAction()

			actions = append(actions, action)
		}

		var setRequestHeader []frontendRuleActionSetHeaderModel
		diags = planAction.SetRequestHeader.ElementsAs(ctx, &setRequestHeader, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, setHeader := range setRequestHeader {
			action := request.NewLoadBalancerSetRequestHeaderAction(setHeader.Header.ValueString(), setHeader.Value.ValueString())

			actions = append(actions, action)
		}

		var setResponseHeader []frontendRuleActionSetHeaderModel
		diags = planAction.SetResponseHeader.ElementsAs(ctx, &setResponseHeader, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, setHeader := range setResponseHeader {
			action := request.NewLoadBalancerSetResponseHeaderAction(setHeader.Header.ValueString(), setHeader.Value.ValueString())

			actions = append(actions, action)
		}

		var httpRedirects []frontendRuleActionHTTPRedirectModel
		diags = planAction.HTTPRedirect.ElementsAs(ctx, &httpRedirects, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, httpRedirect := range httpRedirects {
			var action upcloud.LoadBalancerAction
			if httpRedirect.Scheme.ValueString() != "" {
				action = request.NewLoadBalancerHTTPRedirectSchemeActionWithStatus(upcloud.LoadBalancerActionHTTPRedirectScheme(httpRedirect.Scheme.ValueString()), int(httpRedirect.Status.ValueInt64()))
			} else if httpRedirect.Location.ValueString() != "" {
				action = request.NewLoadBalancerHTTPRedirectActionWithStatus(httpRedirect.Location.ValueString(), int(httpRedirect.Status.ValueInt64()))
			}

			actions = append(actions, action)
		}

		var httpReturns []frontendRuleActionHTTPReturnModel
		diags = planAction.HTTPReturn.ElementsAs(ctx, &httpReturns, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, httpReturn := range httpReturns {
			action := request.NewLoadBalancerHTTPReturnAction(
				int(httpReturn.Status.ValueInt64()),
				httpReturn.ContentType.ValueString(),
				httpReturn.Payload.ValueString(),
			)

			actions = append(actions, action)
		}

		var tcpRejects []frontendRuleActionTCPRejectModel
		diags = planAction.TCPReject.ElementsAs(ctx, &tcpRejects, false)
		if diags.HasError() {
			return nil, diags
		}

		for range tcpRejects {
			action := request.NewLoadBalancerTCPRejectAction()

			actions = append(actions, action)
		}

		var useBackends []frontendRuleActionUseBackendModel
		diags = planAction.UseBackend.ElementsAs(ctx, &useBackends, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, useBackend := range useBackends {
			action := request.NewLoadBalancerUseBackendAction(useBackend.BackendName.ValueString())

			actions = append(actions, action)
		}
	}

	return actions, diags
}

func setFrontendRuleActionsValues(ctx context.Context, data *frontendRuleModel, frontendRule *upcloud.LoadBalancerFrontendRule, blocks map[string]schema.ListNestedBlock) diag.Diagnostics {
	if frontendRule == nil || len(frontendRule.Actions) == 0 {
		return nil
	}

	var diags, respDiagnostics diag.Diagnostics

	elementTypes, err := elementTypesByKey("actions", blocks)
	if err != nil {
		respDiagnostics.AddError("cannot set frontend rule actions", err.Error())

		return respDiagnostics
	}

	httpRedirects := make([]frontendRuleActionHTTPRedirectModel, 0)
	httpReturns := make([]frontendRuleActionHTTPReturnModel, 0)
	setForwardedHeaders := make([]frontendRuleActionSetForwardedHeadersModel, 0)
	setRequestHeader := make([]frontendRuleActionSetHeaderModel, 0)
	setResponseHeader := make([]frontendRuleActionSetHeaderModel, 0)
	tcpRejects := make([]frontendRuleActionTCPRejectModel, 0)
	useBackends := make([]frontendRuleActionUseBackendModel, 0)

	for _, a := range frontendRule.Actions {
		if a.HTTPRedirect != nil {
			httpRedirect := frontendRuleActionHTTPRedirectModel{}

			if a.HTTPRedirect.Scheme != "" {
				httpRedirect.Scheme = types.StringValue(string(a.HTTPRedirect.Scheme))
			}

			if a.HTTPRedirect.Location != "" {
				httpRedirect.Location = types.StringValue(a.HTTPRedirect.Location)
			}

			httpRedirect.Status = types.Int64Value(int64(a.HTTPRedirect.Status))

			httpRedirects = append(httpRedirects, httpRedirect)
		}

		if a.HTTPReturn != nil {
			httpReturns = append(httpReturns, frontendRuleActionHTTPReturnModel{
				ContentType: types.StringValue(a.HTTPReturn.ContentType),
				Status:      types.Int64Value(int64(a.HTTPReturn.Status)),
				Payload:     types.StringValue(a.HTTPReturn.Payload),
			})
		}

		if a.SetForwardedHeaders != nil {
			setForwardedHeaders = append(setForwardedHeaders, frontendRuleActionSetForwardedHeadersModel{
				Active: types.BoolValue(true),
			})
		}

		if a.SetRequestHeader != nil {
			setRequestHeader = append(setRequestHeader, frontendRuleActionSetHeaderModel{
				Header: types.StringValue(a.SetRequestHeader.Header),
				Value:  types.StringValue(a.SetRequestHeader.Value),
			})
		}

		if a.SetResponseHeader != nil {
			setResponseHeader = append(setResponseHeader, frontendRuleActionSetHeaderModel{
				Header: types.StringValue(a.SetResponseHeader.Header),
				Value:  types.StringValue(a.SetResponseHeader.Value),
			})
		}

		if a.TCPReject != nil {
			tcpRejects = append(tcpRejects, frontendRuleActionTCPRejectModel{
				Active: types.BoolValue(true),
			})
		}

		if a.UseBackend != nil {
			useBackends = append(useBackends, frontendRuleActionUseBackendModel{
				BackendName: types.StringValue(a.UseBackend.Backend),
			})
		}
	}

	action := frontendRuleActionModel{}

	if elementType, ok := elementTypes["http_redirect"]; ok {
		action.HTTPRedirect, diags = types.ListValueFrom(ctx, elementType, httpRedirects)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "http_redirect element type not found")
	}

	if elementType, ok := elementTypes["http_return"]; ok {
		action.HTTPReturn, diags = types.ListValueFrom(ctx, elementType, httpReturns)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "http_return element type not found")
	}

	if elementType, ok := elementTypes["set_forwarded_headers"]; ok {
		action.SetForwardedHeaders, diags = types.ListValueFrom(ctx, elementType, setForwardedHeaders)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "set_forwarded_headers element type not found")
	}

	if elementType, ok := elementTypes["set_request_header"]; ok {
		action.SetRequestHeader, diags = types.ListValueFrom(ctx, elementType, setRequestHeader)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "set_request_header element type not found")
	}

	if elementType, ok := elementTypes["set_response_header"]; ok {
		action.SetResponseHeader, diags = types.ListValueFrom(ctx, elementType, setResponseHeader)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "set_response_header element type not found")
	}

	if elementType, ok := elementTypes["tcp_reject"]; ok {
		action.TCPReject, diags = types.ListValueFrom(ctx, elementType, tcpRejects)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "tcp_reject element type not found")
	}

	if elementType, ok := elementTypes["use_backend"]; ok {
		action.UseBackend, diags = types.ListValueFrom(ctx, elementType, useBackends)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule actions", "use_backend element type not found")
	}

	data.Actions, diags = types.ListValueFrom(ctx, data.Actions.ElementType(ctx), []frontendRuleActionModel{action})
	respDiagnostics.Append(diags...)

	return respDiagnostics
}
