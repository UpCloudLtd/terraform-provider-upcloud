package loadbalancer

import (
	"fmt"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		"range_start": {
			Description: "Integer value.",
			Type:        schema.TypeInt,
			Default:     0,
			Optional:    true,
			ForceNew:    true,
		},
		"range_end": {
			Description: "Integer value.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},
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
	}
}

func loadBalancerMatchersFromResourceData(d *schema.ResourceData) ([]upcloud.LoadBalancerMatcher, error) {
	m := make([]upcloud.LoadBalancerMatcher, 0)
	if _, ok := d.GetOk("matchers.0"); !ok {
		return m, nil
	}
	for _, v := range d.Get("matchers.0.src_port").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeSrcPort,
			SrcPort: &upcloud.LoadBalancerMatcherInteger{
				Method:     upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
				Value:      v["value"].(int),
				RangeStart: v["range_start"].(int),
				RangeEnd:   v["range_start"].(int),
			},
		})
	}

	for _, v := range d.Get("matchers.0.src_ip").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeSrcIP,
			SrcIP: &upcloud.LoadBalancerMatcherSourceIP{
				Value: v["value"].(string),
			},
		})
	}

	for _, v := range d.Get("matchers.0.body_size").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeBodySize,
			BodySize: &upcloud.LoadBalancerMatcherInteger{
				Method:     upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
				Value:      v["value"].(int),
				RangeStart: v["range_start"].(int),
				RangeEnd:   v["range_start"].(int),
			},
		})
	}

	for _, v := range d.Get("matchers.0.path").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypePath,
			Path: &upcloud.LoadBalancerMatcherString{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.url").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeURL,
			URL: &upcloud.LoadBalancerMatcherString{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.url_query").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeURLQuery,
			URLQuery: &upcloud.LoadBalancerMatcherString{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.host").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeHost,
			Host: &upcloud.LoadBalancerMatcherHost{
				Value: v["value"].(string),
			},
		})
	}

	for _, v := range d.Get("matchers.0.http_method").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeHTTPMethod,
			HTTPMethod: &upcloud.LoadBalancerMatcherHTTPMethod{
				Value: upcloud.LoadBalancerHTTPMatcherMethod(v["value"].(string)),
			},
		})
	}

	for _, v := range d.Get("matchers.0.cookie").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeCookie,
			Cookie: &upcloud.LoadBalancerMatcherStringWithArgument{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				Name:       v["name"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.header").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeHeader,
			Header: &upcloud.LoadBalancerMatcherStringWithArgument{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				Name:       v["name"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.url_param").([]interface{}) {
		v := v.(map[string]interface{})
		ic := v["ignore_case"].(bool)
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeURLParam,
			URLParam: &upcloud.LoadBalancerMatcherStringWithArgument{
				Method:     upcloud.LoadBalancerStringMatcherMethod(v["method"].(string)),
				Value:      v["value"].(string),
				Name:       v["name"].(string),
				IgnoreCase: &ic,
			},
		})
	}

	for _, v := range d.Get("matchers.0.num_members_up").([]interface{}) {
		v := v.(map[string]interface{})
		m = append(m, upcloud.LoadBalancerMatcher{
			Type: upcloud.LoadBalancerMatcherTypeNumMembersUP,
			NumMembersUP: &upcloud.LoadBalancerMatcherNumMembersUP{
				Method:  upcloud.LoadBalancerIntegerMatcherMethod(v["method"].(string)),
				Value:   v["value"].(int),
				Backend: v["backend_name"].(string),
			},
		})
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
				"value": m.SrcIP.Value,
			}
		case upcloud.LoadBalancerMatcherTypeSrcPort:
			v = map[string]interface{}{
				"method":      m.SrcPort.Method,
				"value":       m.SrcPort.Value,
				"range_start": m.SrcPort.RangeStart,
				"range_end":   m.SrcPort.RangeEnd,
			}
		case upcloud.LoadBalancerMatcherTypeBodySize:
			v = map[string]interface{}{
				"method":      m.BodySize.Method,
				"value":       m.BodySize.Value,
				"range_start": m.BodySize.RangeStart,
				"range_end":   m.BodySize.RangeEnd,
			}
		case upcloud.LoadBalancerMatcherTypePath:
			v = map[string]interface{}{
				"value":       m.Path.Value,
				"ignore_case": m.Path.IgnoreCase,
				"method":      m.Path.Method,
			}
		case upcloud.LoadBalancerMatcherTypeURL:
			v = map[string]interface{}{
				"value":       m.URL.Value,
				"ignore_case": m.URL.IgnoreCase,
				"method":      m.URL.Method,
			}
		case upcloud.LoadBalancerMatcherTypeURLParam:
			v = map[string]interface{}{
				"value":       m.URLParam.Value,
				"ignore_case": m.URLParam.IgnoreCase,
				"name":        m.URLParam.Name,
				"method":      m.URLParam.Method,
			}
		case upcloud.LoadBalancerMatcherTypeURLQuery:
			v = map[string]interface{}{
				"value":       m.URLQuery.Value,
				"ignore_case": m.URLQuery.IgnoreCase,
				"method":      m.URLQuery.Method,
			}
		case upcloud.LoadBalancerMatcherTypeHost:
			v = map[string]interface{}{
				"value": m.Host.Value,
			}
		case upcloud.LoadBalancerMatcherTypeHTTPMethod:
			v = map[string]interface{}{
				"value": m.HTTPMethod.Value,
			}
		case upcloud.LoadBalancerMatcherTypeCookie:
			v = map[string]interface{}{
				"value":       m.Cookie.Value,
				"name":        m.Cookie.Name,
				"method":      m.Cookie.Method,
				"ignore_case": m.Cookie.IgnoreCase,
			}
		case upcloud.LoadBalancerMatcherTypeHeader:
			v = map[string]interface{}{
				"value":       m.Header.Value,
				"name":        m.Header.Name,
				"method":      m.Header.Method,
				"ignore_case": m.Header.IgnoreCase,
			}
		case upcloud.LoadBalancerMatcherTypeNumMembersUP:
			v = map[string]interface{}{
				"value":        m.NumMembersUP.Value,
				"method":       m.NumMembersUP.Method,
				"backend_name": m.NumMembersUP.Backend,
			}
		default:
			return fmt.Errorf("received unsupported matcher type '%s' %+v", m.Type, m)
		}
		if v != nil {
			matchers[t] = append(matchers[t], v)
		}
	}
	return d.Set("matchers", []interface{}{matchers})
}
