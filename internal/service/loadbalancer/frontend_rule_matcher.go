package loadbalancer

import (
	"context"
	"fmt"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type frontendRuleMatcherModel struct {
	BodySize      types.List `tfsdk:"body_size"`
	BodySizeRange types.List `tfsdk:"body_size_range"`
	Cookie        types.List `tfsdk:"cookie"`
	Header        types.List `tfsdk:"header"`
	Host          types.List `tfsdk:"host"`
	HTTPMethod    types.List `tfsdk:"http_method"`
	NumMembersUp  types.List `tfsdk:"num_members_up"`
	Path          types.List `tfsdk:"path"`
	SrcIP         types.List `tfsdk:"src_ip"`
	SrcPort       types.List `tfsdk:"src_port"`
	SrcPortRange  types.List `tfsdk:"src_port_range"`
	URL           types.List `tfsdk:"url"`
	URLParam      types.List `tfsdk:"url_param"`
	URLQuery      types.List `tfsdk:"url_query"`
}

type frontendRuleMatcherHostModel struct {
	Value   types.String `tfsdk:"value"`
	Inverse types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherHTTPMethodModel struct {
	Value   types.String `tfsdk:"value"`
	Inverse types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherNumMembersUpModel struct {
	Method      types.String `tfsdk:"method"`
	Value       types.Int64  `tfsdk:"value"`
	BackendName types.String `tfsdk:"backend_name"`
	Inverse     types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherSrcIPModel struct {
	Value   types.String `tfsdk:"value"`
	Inverse types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherIntegerModel struct {
	Method  types.String `tfsdk:"method"`
	Value   types.Int64  `tfsdk:"value"`
	Inverse types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherRangeModel struct {
	RangeStart types.Int64 `tfsdk:"range_start"`
	RangeEnd   types.Int64 `tfsdk:"range_end"`
	Inverse    types.Bool  `tfsdk:"inverse"`
}

type frontendRuleMatcherStringWithArgumentModel struct {
	Method     types.String `tfsdk:"method"`
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	IgnoreCase types.Bool   `tfsdk:"ignore_case"`
	Inverse    types.Bool   `tfsdk:"inverse"`
}

type frontendRuleMatcherStringModel struct {
	Method     types.String `tfsdk:"method"`
	Value      types.String `tfsdk:"value"`
	IgnoreCase types.Bool   `tfsdk:"ignore_case"`
	Inverse    types.Bool   `tfsdk:"inverse"`
}

func frontendRuleMatcherHTTPMethodSchema() schema.NestedBlockObject {
	methods := []string{
		string(upcloud.LoadBalancerHTTPMatcherMethodGet),
		string(upcloud.LoadBalancerHTTPMatcherMethodHead),
		string(upcloud.LoadBalancerHTTPMatcherMethodPost),
		string(upcloud.LoadBalancerHTTPMatcherMethodPut),
		string(upcloud.LoadBalancerHTTPMatcherMethodPatch),
		string(upcloud.LoadBalancerHTTPMatcherMethodDelete),
		string(upcloud.LoadBalancerHTTPMatcherMethodConnect),
		string(upcloud.LoadBalancerHTTPMatcherMethodOptions),
		string(upcloud.LoadBalancerHTTPMatcherMethodTrace),
	}

	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("String value (`%s`).", strings.Join(methods, "`, `")),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(methods...),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherBackendSchema() schema.NestedBlockObject {
	methods := []string{
		string(upcloud.LoadBalancerIntegerMatcherMethodEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreater),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreaterOrEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodLess),
		string(upcloud.LoadBalancerIntegerMatcherMethodLessOrEqual),
	}

	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Match method (`%s`).", strings.Join(methods, "`, `")),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(methods...),
				},
			},
			"value": schema.Int64Attribute{
				MarkdownDescription: "Integer value.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"backend_name": schema.StringAttribute{
				MarkdownDescription: "The name of the `backend`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherStringWithArgumentSchema() schema.NestedBlockObject {
	methods := []string{
		string(upcloud.LoadBalancerStringMatcherMethodExact),
		string(upcloud.LoadBalancerStringMatcherMethodSubstring),
		string(upcloud.LoadBalancerStringMatcherMethodRegexp),
		string(upcloud.LoadBalancerStringMatcherMethodStarts),
		string(upcloud.LoadBalancerStringMatcherMethodEnds),
		string(upcloud.LoadBalancerStringMatcherMethodDomain),
		string(upcloud.LoadBalancerStringMatcherMethodIP),
		string(upcloud.LoadBalancerStringMatcherMethodExists),
	}

	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Match method (`%s`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.", strings.Join(methods, "`, `")),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(methods...),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the argument.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "String value.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"ignore_case": frontendRuleMatcherIgnoreCaseSchema(),
			"inverse":     frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherIPSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{
				MarkdownDescription: "IP address. CIDR masks are supported, e.g. `192.168.0.0/24`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherStringSchema() schema.NestedBlockObject {
	methods := []string{
		string(upcloud.LoadBalancerStringMatcherMethodExact),
		string(upcloud.LoadBalancerStringMatcherMethodSubstring),
		string(upcloud.LoadBalancerStringMatcherMethodRegexp),
		string(upcloud.LoadBalancerStringMatcherMethodStarts),
		string(upcloud.LoadBalancerStringMatcherMethodEnds),
		string(upcloud.LoadBalancerStringMatcherMethodDomain),
		string(upcloud.LoadBalancerStringMatcherMethodIP),
		string(upcloud.LoadBalancerStringMatcherMethodExists),
	}

	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				Description: fmt.Sprintf("Match method (`%s`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.", strings.Join(methods, "`, `")),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(methods...),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "String value.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"ignore_case": frontendRuleMatcherIgnoreCaseSchema(),
			"inverse":     frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherHostSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{
				MarkdownDescription: "String value.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherIntegerSchema() schema.NestedBlockObject {
	methods := []string{
		string(upcloud.LoadBalancerIntegerMatcherMethodEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreater),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreaterOrEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodLess),
		string(upcloud.LoadBalancerIntegerMatcherMethodLessOrEqual),
	}

	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"method": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Match method (`%s`).", strings.Join(methods, "`, `")),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(methods...),
				},
			},
			"value": schema.Int64Attribute{
				MarkdownDescription: "Integer value.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherRangeSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"range_start": schema.Int64Attribute{
				MarkdownDescription: "Integer value.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"range_end": schema.Int64Attribute{
				MarkdownDescription: "Integer value.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"inverse": frontendRuleMatcherInverseSchema(),
		},
	}
}

func frontendRuleMatcherInverseSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Defines if the condition should be inverted. Works similarly to logical NOT operator.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
}

func frontendRuleMatcherIgnoreCaseSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Defines if case should be ignored. Defaults to `false`.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.RequiresReplace(),
		},
	}
}

func buildFrontendRuleMatchers(ctx context.Context, dataMatchers types.List) ([]upcloud.LoadBalancerMatcher, diag.Diagnostics) {
	if dataMatchers.IsNull() {
		return nil, nil
	}

	var planMatchers []frontendRuleMatcherModel
	diags := dataMatchers.ElementsAs(ctx, &planMatchers, false)
	if diags.HasError() {
		return nil, diags
	}

	matchers := make([]upcloud.LoadBalancerMatcher, 0)
	for _, planMatcher := range planMatchers {
		var bodySizes []frontendRuleMatcherIntegerModel
		diags = planMatcher.BodySize.ElementsAs(ctx, &bodySizes, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, bodySize := range bodySizes {
			matcher := request.NewLoadBalancerBodySizeMatcher(
				upcloud.LoadBalancerIntegerMatcherMethod(bodySize.Method.ValueString()),
				int(bodySize.Value.ValueInt64()),
			)
			matchers = appendMatcher(matchers, matcher, bodySize.Inverse.ValueBool())
		}

		var bodySizeRanges []frontendRuleMatcherRangeModel
		diags = planMatcher.BodySizeRange.ElementsAs(ctx, &bodySizeRanges, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, bodySizeRange := range bodySizeRanges {
			matcher := request.NewLoadBalancerBodySizeRangeMatcher(
				int(bodySizeRange.RangeStart.ValueInt64()),
				int(bodySizeRange.RangeEnd.ValueInt64()),
			)
			matchers = appendMatcher(matchers, matcher, bodySizeRange.Inverse.ValueBool())
		}

		var cookies []frontendRuleMatcherStringWithArgumentModel
		diags = planMatcher.Cookie.ElementsAs(ctx, &cookies, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, cookie := range cookies {
			matcher := request.NewLoadBalancerCookieMatcher(
				upcloud.LoadBalancerStringMatcherMethod(cookie.Method.ValueString()),
				cookie.Name.ValueString(),
				cookie.Value.ValueString(),
				cookie.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, cookie.Inverse.ValueBool())
		}

		var headers []frontendRuleMatcherStringWithArgumentModel
		diags = planMatcher.Header.ElementsAs(ctx, &headers, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, header := range headers {
			matcher := request.NewLoadBalancerHeaderMatcher(
				upcloud.LoadBalancerStringMatcherMethod(header.Method.ValueString()),
				header.Name.ValueString(),
				header.Value.ValueString(),
				header.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, header.Inverse.ValueBool())
		}

		var hosts []frontendRuleMatcherHostModel
		diags = planMatcher.Host.ElementsAs(ctx, &hosts, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, host := range hosts {
			matcher := request.NewLoadBalancerHostMatcher(
				host.Value.ValueString(),
			)
			matchers = appendMatcher(matchers, matcher, host.Inverse.ValueBool())
		}

		var httpMethods []frontendRuleMatcherHTTPMethodModel
		diags = planMatcher.HTTPMethod.ElementsAs(ctx, &httpMethods, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, httpMethod := range httpMethods {
			matcher := request.NewLoadBalancerHTTPMethodMatcher(
				upcloud.LoadBalancerHTTPMatcherMethod(httpMethod.Value.ValueString()),
			)
			matchers = appendMatcher(matchers, matcher, httpMethod.Inverse.ValueBool())
		}

		var numMembersUp []frontendRuleMatcherNumMembersUpModel
		diags = planMatcher.NumMembersUp.ElementsAs(ctx, &numMembersUp, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, numMembers := range numMembersUp {
			matcher := request.NewLoadBalancerNumMembersUpMatcher(
				upcloud.LoadBalancerIntegerMatcherMethod(numMembers.Method.ValueString()),
				int(numMembers.Value.ValueInt64()),
				numMembers.BackendName.ValueString(),
			)
			matchers = appendMatcher(matchers, matcher, numMembers.Inverse.ValueBool())
		}

		var paths []frontendRuleMatcherStringModel
		diags = planMatcher.Path.ElementsAs(ctx, &paths, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, pathMatcher := range paths {
			matcher := request.NewLoadBalancerPathMatcher(
				upcloud.LoadBalancerStringMatcherMethod(pathMatcher.Method.ValueString()),
				pathMatcher.Value.ValueString(),
				pathMatcher.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, pathMatcher.Inverse.ValueBool())
		}

		var srcIPs []frontendRuleMatcherSrcIPModel
		diags = planMatcher.SrcIP.ElementsAs(ctx, &srcIPs, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, srcIP := range srcIPs {
			matcher := request.NewLoadBalancerSrcIPMatcher(
				srcIP.Value.ValueString(),
			)
			matchers = appendMatcher(matchers, matcher, srcIP.Inverse.ValueBool())
		}

		var srcPorts []frontendRuleMatcherIntegerModel
		diags = planMatcher.SrcPort.ElementsAs(ctx, &srcPorts, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, srcPort := range srcPorts {
			matcher := request.NewLoadBalancerSrcPortMatcher(
				upcloud.LoadBalancerIntegerMatcherMethod(srcPort.Method.ValueString()),
				int(srcPort.Value.ValueInt64()),
			)
			matchers = appendMatcher(matchers, matcher, srcPort.Inverse.ValueBool())
		}

		var srcPortRanges []frontendRuleMatcherRangeModel
		diags = planMatcher.SrcPortRange.ElementsAs(ctx, &srcPortRanges, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, srcPortRange := range srcPortRanges {
			matcher := request.NewLoadBalancerSrcPortRangeMatcher(
				int(srcPortRange.RangeStart.ValueInt64()),
				int(srcPortRange.RangeEnd.ValueInt64()),
			)
			matchers = appendMatcher(matchers, matcher, srcPortRange.Inverse.ValueBool())
		}

		var urls []frontendRuleMatcherStringModel
		diags = planMatcher.URL.ElementsAs(ctx, &urls, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, url := range urls {
			matcher := request.NewLoadBalancerURLMatcher(
				upcloud.LoadBalancerStringMatcherMethod(url.Method.ValueString()),
				url.Value.ValueString(),
				url.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, url.Inverse.ValueBool())
		}

		var urlParams []frontendRuleMatcherStringWithArgumentModel
		diags = planMatcher.URLParam.ElementsAs(ctx, &urlParams, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, urlParam := range urlParams {
			matcher := request.NewLoadBalancerURLParamMatcher(
				upcloud.LoadBalancerStringMatcherMethod(urlParam.Method.ValueString()),
				urlParam.Name.ValueString(),
				urlParam.Value.ValueString(),
				urlParam.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, urlParam.Inverse.ValueBool())
		}

		var urlQueries []frontendRuleMatcherStringModel
		diags = planMatcher.URLQuery.ElementsAs(ctx, &urlQueries, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, urlQuery := range urlQueries {
			matcher := request.NewLoadBalancerURLQueryMatcher(
				upcloud.LoadBalancerStringMatcherMethod(urlQuery.Method.ValueString()),
				urlQuery.Value.ValueString(),
				urlQuery.IgnoreCase.ValueBoolPointer(),
			)
			matchers = appendMatcher(matchers, matcher, urlQuery.Inverse.ValueBool())
		}
	}

	return matchers, diags
}

func setFrontendRuleMatchersValues(ctx context.Context, data *frontendRuleModel, frontendRule *upcloud.LoadBalancerFrontendRule, blocks map[string]schema.ListNestedBlock) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	elementTypes, err := elementTypesByKey("matchers", blocks)
	if err != nil {
		respDiagnostics.AddError("cannot set frontend rule matchers", err.Error())

		return respDiagnostics
	}

	bodySizes := make([]frontendRuleMatcherIntegerModel, 0)
	bodySizeRanges := make([]frontendRuleMatcherRangeModel, 0)
	cookies := make([]frontendRuleMatcherStringWithArgumentModel, 0)
	headers := make([]frontendRuleMatcherStringWithArgumentModel, 0)
	hosts := make([]frontendRuleMatcherHostModel, 0)
	httpMethods := make([]frontendRuleMatcherHTTPMethodModel, 0)
	numMembersUp := make([]frontendRuleMatcherNumMembersUpModel, 0)
	paths := make([]frontendRuleMatcherStringModel, 0)
	srcIPs := make([]frontendRuleMatcherSrcIPModel, 0)
	srcPorts := make([]frontendRuleMatcherIntegerModel, 0)
	srcPortRanges := make([]frontendRuleMatcherRangeModel, 0)
	urls := make([]frontendRuleMatcherStringModel, 0)
	urlParams := make([]frontendRuleMatcherStringWithArgumentModel, 0)
	urlQueries := make([]frontendRuleMatcherStringModel, 0)

	for _, m := range frontendRule.Matchers {
		if m.BodySize != nil {
			if m.BodySize.Method == upcloud.LoadBalancerIntegerMatcherMethodRange {
				bodySizeRanges = append(bodySizeRanges, frontendRuleMatcherRangeModel{
					Inverse:    types.BoolValue(*m.Inverse),
					RangeEnd:   types.Int64Value(int64(m.BodySize.RangeEnd)),
					RangeStart: types.Int64Value(int64(m.BodySize.RangeStart)),
				})
			} else {
				bodySizes = append(bodySizes, frontendRuleMatcherIntegerModel{
					Method:  types.StringValue(string(m.BodySize.Method)),
					Inverse: types.BoolValue(*m.Inverse),
					Value:   types.Int64Value(int64(m.BodySize.Value)),
				})
			}
		}

		if m.Cookie != nil {
			var ignoreCase bool
			if m.Cookie.IgnoreCase != nil {
				ignoreCase = *m.Cookie.IgnoreCase
			}
			cookies = append(cookies, frontendRuleMatcherStringWithArgumentModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.Cookie.Method)),
				Name:       types.StringValue(m.Cookie.Name),
				Value:      types.StringValue(m.Cookie.Value),
			})
		}

		if m.Header != nil {
			var ignoreCase bool
			if m.Header.IgnoreCase != nil {
				ignoreCase = *m.Header.IgnoreCase
			}
			headers = append(headers, frontendRuleMatcherStringWithArgumentModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.Header.Method)),
				Name:       types.StringValue(m.Header.Name),
				Value:      types.StringValue(m.Header.Value),
			})
		}

		if m.Host != nil {
			hosts = append(hosts, frontendRuleMatcherHostModel{
				Inverse: types.BoolValue(*m.Inverse),
				Value:   types.StringValue(m.Host.Value),
			})
		}

		if m.HTTPMethod != nil {
			httpMethods = append(httpMethods, frontendRuleMatcherHTTPMethodModel{
				Inverse: types.BoolValue(*m.Inverse),
				Value:   types.StringValue(string(m.HTTPMethod.Value)),
			})
		}

		if m.NumMembersUp != nil {
			numMembersUp = append(numMembersUp, frontendRuleMatcherNumMembersUpModel{
				BackendName: types.StringValue(m.NumMembersUp.Backend),
				Inverse:     types.BoolValue(*m.Inverse),
				Method:      types.StringValue(string(m.NumMembersUp.Method)),
				Value:       types.Int64Value(int64(m.NumMembersUp.Value)),
			})
		}

		if m.Path != nil {
			var ignoreCase bool
			if m.Path.IgnoreCase != nil {
				ignoreCase = *m.Path.IgnoreCase
			}
			paths = append(paths, frontendRuleMatcherStringModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.Path.Method)),
				Value:      types.StringValue(m.Path.Value),
			})
		}

		if m.SrcIP != nil {
			srcIPs = append(srcIPs, frontendRuleMatcherSrcIPModel{
				Inverse: types.BoolValue(*m.Inverse),
				Value:   types.StringValue(m.SrcIP.Value),
			})
		}

		if m.SrcPort != nil {
			if m.SrcPort.Method == upcloud.LoadBalancerIntegerMatcherMethodRange {
				srcPortRanges = append(srcPortRanges, frontendRuleMatcherRangeModel{
					Inverse:    types.BoolValue(*m.Inverse),
					RangeEnd:   types.Int64Value(int64(m.SrcPort.RangeEnd)),
					RangeStart: types.Int64Value(int64(m.SrcPort.RangeStart)),
				})
			} else {
				srcPorts = append(srcPorts, frontendRuleMatcherIntegerModel{
					Inverse: types.BoolValue(*m.Inverse),
					Method:  types.StringValue(string(m.SrcPort.Method)),
					Value:   types.Int64Value(int64(m.SrcPort.Value)),
				})
			}
		}

		if m.URL != nil {
			var ignoreCase bool
			if m.URL.IgnoreCase != nil {
				ignoreCase = *m.URL.IgnoreCase
			}
			urls = append(urls, frontendRuleMatcherStringModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.URL.Method)),
				Value:      types.StringValue(m.URL.Value),
			})
		}

		if m.URLParam != nil {
			var ignoreCase bool
			if m.URLParam.IgnoreCase != nil {
				ignoreCase = *m.URLParam.IgnoreCase
			}
			urlParams = append(urlParams, frontendRuleMatcherStringWithArgumentModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.URLParam.Method)),
				Name:       types.StringValue(m.URLParam.Name),
				Value:      types.StringValue(m.URLParam.Value),
			})
		}

		if m.URLQuery != nil {
			var ignoreCase bool
			if m.URLQuery.IgnoreCase != nil {
				ignoreCase = *m.URLQuery.IgnoreCase
			}
			urlQueries = append(urlQueries, frontendRuleMatcherStringModel{
				IgnoreCase: types.BoolValue(ignoreCase),
				Inverse:    types.BoolValue(*m.Inverse),
				Method:     types.StringValue(string(m.URLQuery.Method)),
				Value:      types.StringValue(m.URLQuery.Value),
			})
		}
	}

	matcher := frontendRuleMatcherModel{}

	if elementType, ok := elementTypes["body_size"]; ok {
		matcher.BodySize, diags = types.ListValueFrom(ctx, elementType, bodySizes)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "body_size element type not found")
	}

	if elementType, ok := elementTypes["body_size_range"]; ok {
		matcher.BodySizeRange, diags = types.ListValueFrom(ctx, elementType, bodySizeRanges)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "body_size_range element type not found")
	}

	if elementType, ok := elementTypes["cookie"]; ok {
		matcher.Cookie, diags = types.ListValueFrom(ctx, elementType, cookies)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "cookie element type not found")
	}

	if elementType, ok := elementTypes["header"]; ok {
		matcher.Header, diags = types.ListValueFrom(ctx, elementType, headers)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "header element type not found")
	}

	if elementType, ok := elementTypes["host"]; ok {
		matcher.Host, diags = types.ListValueFrom(ctx, elementType, hosts)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "host element type not found")
	}

	if elementType, ok := elementTypes["http_method"]; ok {
		matcher.HTTPMethod, diags = types.ListValueFrom(ctx, elementType, httpMethods)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "http_method element type not found")
	}

	if elementType, ok := elementTypes["num_members_up"]; ok {
		matcher.NumMembersUp, diags = types.ListValueFrom(ctx, elementType, numMembersUp)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "num_members_up element type not found")
	}

	if elementType, ok := elementTypes["path"]; ok {
		matcher.Path, diags = types.ListValueFrom(ctx, elementType, paths)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "path element type not found")
	}

	if elementType, ok := elementTypes["src_ip"]; ok {
		matcher.SrcIP, diags = types.ListValueFrom(ctx, elementType, srcIPs)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "src_ip element type not found")
	}

	if elementType, ok := elementTypes["src_port"]; ok {
		matcher.SrcPort, diags = types.ListValueFrom(ctx, elementType, srcPorts)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "src_port element type not found")
	}

	if elementType, ok := elementTypes["src_port_range"]; ok {
		matcher.SrcPortRange, diags = types.ListValueFrom(ctx, elementType, srcPortRanges)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "src_port_range element type not found")
	}

	if elementType, ok := elementTypes["url"]; ok {
		matcher.URL, diags = types.ListValueFrom(ctx, elementType, urls)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "url element type not found")
	}

	if elementType, ok := elementTypes["url_param"]; ok {
		matcher.URLParam, diags = types.ListValueFrom(ctx, elementType, urlParams)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "url_param element type not found")
	}

	if elementType, ok := elementTypes["url_query"]; ok {
		matcher.URLQuery, diags = types.ListValueFrom(ctx, elementType, urlQueries)
		respDiagnostics.Append(diags...)
	} else {
		respDiagnostics.AddError("cannot set frontend rule matcher", "url_query element type not found")
	}

	data.Matchers, diags = types.ListValueFrom(ctx, data.Matchers.ElementType(ctx), []frontendRuleMatcherModel{matcher})
	respDiagnostics.Append(diags...)

	return respDiagnostics
}

func appendMatcher(
	matchers []upcloud.LoadBalancerMatcher,
	newMatcher upcloud.LoadBalancerMatcher,
	inverse interface{},
) []upcloud.LoadBalancerMatcher {
	if inverse.(bool) {
		return append(matchers, request.NewLoadBalancerInverseMatcher(newMatcher))
	}
	return append(matchers, newMatcher)
}
