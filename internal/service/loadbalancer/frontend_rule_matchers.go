package loadbalancer

import (
	"fmt"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	customSrcPortRangeMatcherType  = "src_port_range"
	customBodySizeRangeMatcherType = "body_size_range"
)

func frontendRuleMatchersSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"src_port": {
			Description: "Matches by source port number.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherIntegerSchema(),
			},
		},
		"src_port_range": {
			Description: "Matches by range of source port numbers",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherRangeSchema(),
			},
		},
		"src_ip": {
			Description: "Matches by source IP address.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherIPSchema(),
			},
		},
		"body_size": {
			Description: "Matches by HTTP request body size.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherIntegerSchema(),
			},
		},
		"body_size_range": {
			Description: "Matches by range of HTTP request body sizes",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherRangeSchema(),
			},
		},
		"path": {
			Description: "Matches by URL path.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringSchema(),
			},
		},
		"url": {
			Description: "Matches by URL without schema, e.g. `example.com/dashboard`.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringSchema(),
			},
		},
		"url_query": {
			Description: "Matches by URL query string.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringSchema(),
			},
		},
		"host": {
			Description: "Matches by hostname. Header extracted from HTTP Headers or from TLS certificate in case of secured connection.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherHostSchema(),
			},
		},
		"http_method": {
			Description: "Matches by HTTP method.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherHTTPMethodSchema(),
			},
		},
		"cookie": {
			Description: "Matches by HTTP cookie value. Cookie name must be provided.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringWithArgumentSchema(),
			},
		},
		"header": {
			Description: "Matches by HTTP header value. Header name must be provided.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringWithArgumentSchema(),
			},
		},
		"url_param": {
			Description: "Matches by URL query parameter value. Query parameter name must be provided",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherStringWithArgumentSchema(),
			},
		},
		"num_members_up": {
			Description: "Matches by number of healthy backend members.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: frontendRuleMatcherBackendSchema(),
			},
		},
	}
}

func inverseSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Sets if the condition should be inverted. Works similar to logical NOT operator.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	}
}

func frontendRuleMatcherBackendSchema() map[string]*schema.Schema {
	methods := []string{
		string(upcloud.LoadBalancerIntegerMatcherMethodEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreater),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreaterOrEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodLess),
		string(upcloud.LoadBalancerIntegerMatcherMethodLessOrEqual),
	}

	return map[string]*schema.Schema{
		"method": {
			Description:      fmt.Sprintf("Match method (`%s`).", strings.Join(methods, "`, `")),
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(methods, false)),
		},
		"value": {
			Description:      "Integer value.",
			Type:             schema.TypeInt,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
		},
		"backend_name": {
			Description:      "The name of the `backend` which members will be monitored.",
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherStringWithArgumentSchema() map[string]*schema.Schema {
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

	return map[string]*schema.Schema{
		"method": {
			Description:      fmt.Sprintf("Match method (`%s`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.", strings.Join(methods, "`, `")),
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(methods, false)),
		},
		"name": {
			Description:      "Name of the argument.",
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"value": {
			Description:      "String value.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"ignore_case": {
			Description: "Ignore case, default `false`.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true,
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherStringSchema() map[string]*schema.Schema {
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

	return map[string]*schema.Schema{
		"method": {
			Description:      fmt.Sprintf("Match method (`%s`). Matcher with `exists` and `ip` methods must be used without `value` and `ignore_case` fields.", strings.Join(methods, "`, `")),
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(methods, false)),
		},
		"value": {
			Description:      "String value.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"ignore_case": {
			Description: "Ignore case, default `false`.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherHTTPMethodSchema() map[string]*schema.Schema {
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

	return map[string]*schema.Schema{
		"value": {
			Description:      fmt.Sprintf("String value (`%s`).", strings.Join(methods, "`, `")),
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(methods, false)),
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"value": {
			Description:      "String value.",
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherIntegerSchema() map[string]*schema.Schema {
	methods := []string{
		string(upcloud.LoadBalancerIntegerMatcherMethodEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreater),
		string(upcloud.LoadBalancerIntegerMatcherMethodGreaterOrEqual),
		string(upcloud.LoadBalancerIntegerMatcherMethodLess),
		string(upcloud.LoadBalancerIntegerMatcherMethodLessOrEqual),
	}

	return map[string]*schema.Schema{
		"method": {
			Description:      fmt.Sprintf("Match method (`%s`).", strings.Join(methods, "`, `")),
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(methods, false)),
		},
		"value": {
			Description: "Integer value.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherRangeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"range_start": {
			Description: "Integer value.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"range_end": {
			Description: "Integer value.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"inverse": inverseSchema(),
	}
}

func frontendRuleMatcherIPSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"value": {
			Description: "IP address. CIDR masks are supported, e.g. `192.168.0.0/24`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"inverse": inverseSchema(),
	}
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

func loadBalancerMatchersFromResourceData(d *schema.ResourceData) ([]upcloud.LoadBalancerMatcher, error) {
	m := make([]upcloud.LoadBalancerMatcher, 0)
	if _, ok := d.GetOk("matchers.0"); !ok {
		return m, nil
	}
	for _, v := range d.Get("matchers.0.src_port").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerSrcPortMatcher(
			upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
			v["value"].(int),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.src_port_range").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerSrcPortRangeMatcher(
			v["range_start"].(int),
			v["range_end"].(int),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.src_ip").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerSrcIPMatcher(v["value"].(string)), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.body_size").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerBodySizeMatcher(
			upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
			v["value"].(int),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.body_size_range").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerBodySizeRangeMatcher(
			v["range_start"].(int),
			v["range_end"].(int),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.path").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerPathMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.url").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerURLMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.url_query").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerURLQueryMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.host").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerHostMatcher(v["value"].(string)), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.http_method").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerHTTPMethodMatcher(
			upcloud.LoadBalancerHTTPMatcherMethod(v["value"].(string)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.cookie").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerCookieMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["name"].(string),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.header").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerHeaderMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["name"].(string),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.url_param").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerURLParamMatcher(
			upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
			v["name"].(string),
			v["value"].(string),
			upcloud.BoolPtr(v["ignore_case"].(bool)),
		), v["inverse"])
	}

	for _, v := range d.Get("matchers.0.num_members_up").([]interface{}) {
		v := v.(map[string]interface{})
		m = appendMatcher(m, request.NewLoadBalancerNumMembersUpMatcher(
			upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
			v["value"].(int),
			v["backend_name"].(string),
		), v["inverse"])
	}

	return m, nil
}

func setFrontendRuleMatchersResourceData(d *schema.ResourceData, rule *upcloud.LoadBalancerFrontendRule) error {
	if len(rule.Matchers) == 0 {
		return d.Set("matchers", nil)
	}

	matchers := make(map[string][]interface{})
	for _, m := range rule.Matchers {
		t := string(m.Type)
		var v map[string]interface{}
		switch m.Type {
		case upcloud.LoadBalancerMatcherTypeSrcIP:
			v = map[string]interface{}{
				"value":   m.SrcIP.Value,
				"inverse": m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeSrcPort:
			if m.SrcPort.Method == upcloud.LoadBalancerIntegerMatcherMethodRange {
				t = customSrcPortRangeMatcherType

				v = map[string]interface{}{
					"range_start": m.SrcPort.RangeStart,
					"range_end":   m.SrcPort.RangeEnd,
					"inverse":     m.Inverse,
				}
			} else {
				v = map[string]interface{}{
					"method":  m.SrcPort.Method,
					"value":   m.SrcPort.Value,
					"inverse": m.Inverse,
				}
			}
		case upcloud.LoadBalancerMatcherTypeBodySize:
			if m.BodySize.Method == upcloud.LoadBalancerIntegerMatcherMethodRange {
				t = customBodySizeRangeMatcherType

				v = map[string]interface{}{
					"range_start": m.BodySize.RangeStart,
					"range_end":   m.BodySize.RangeEnd,
					"inverse":     m.Inverse,
				}
			} else {
				v = map[string]interface{}{
					"method":  m.BodySize.Method,
					"value":   m.BodySize.Value,
					"inverse": m.Inverse,
				}
			}
		case upcloud.LoadBalancerMatcherTypePath:
			v = map[string]interface{}{
				"value":       m.Path.Value,
				"ignore_case": m.Path.IgnoreCase,
				"method":      m.Path.Method,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeURL:
			v = map[string]interface{}{
				"value":       m.URL.Value,
				"ignore_case": m.URL.IgnoreCase,
				"method":      m.URL.Method,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeURLParam:
			v = map[string]interface{}{
				"value":       m.URLParam.Value,
				"ignore_case": m.URLParam.IgnoreCase,
				"name":        m.URLParam.Name,
				"method":      m.URLParam.Method,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeURLQuery:
			v = map[string]interface{}{
				"value":       m.URLQuery.Value,
				"ignore_case": m.URLQuery.IgnoreCase,
				"method":      m.URLQuery.Method,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeHost:
			v = map[string]interface{}{
				"value":   m.Host.Value,
				"inverse": m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeHTTPMethod:
			v = map[string]interface{}{
				"value":   m.HTTPMethod.Value,
				"inverse": m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeCookie:
			v = map[string]interface{}{
				"value":       m.Cookie.Value,
				"name":        m.Cookie.Name,
				"method":      m.Cookie.Method,
				"ignore_case": m.Cookie.IgnoreCase,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeHeader:
			v = map[string]interface{}{
				"value":       m.Header.Value,
				"name":        m.Header.Name,
				"method":      m.Header.Method,
				"ignore_case": m.Header.IgnoreCase,
				"inverse":     m.Inverse,
			}
		case upcloud.LoadBalancerMatcherTypeNumMembersUp:
			v = map[string]interface{}{
				"value":        m.NumMembersUp.Value,
				"method":       m.NumMembersUp.Method,
				"backend_name": m.NumMembersUp.Backend,
				"inverse":      m.Inverse,
			}
		default:
			return fmt.Errorf("received unsupported matcher type '%s' %+v", m.Type, m)
		}

		matchers[t] = append(matchers[t], v)
	}

	return d.Set("matchers", []interface{}{matchers})
}
