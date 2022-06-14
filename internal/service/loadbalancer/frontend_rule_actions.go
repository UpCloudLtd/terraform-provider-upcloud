package loadbalancer

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func frontendRuleActionsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"use_backend": {
			Description: "Routes traffic to specified `backend`.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"backend_name": {
						Description: "The name of the backend where traffic will be routed.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
				},
			},
		},
		"http_redirect": {
			Description: "Redirects HTTP requests to specified location or URL schema.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"location": {
						Description: "Target location.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
				},
			},
		},
		"http_return": {
			Description: "Returns HTTP response with specified HTTP status.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"content_type": {
						Description: "Content type.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"status": {
						Description:      "HTTP status code.",
						Type:             schema.TypeInt,
						Required:         true,
						ForceNew:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(100, 599)),
					},
					"payload": {
						Description:      "The payload.",
						Type:             schema.TypeString,
						Required:         true,
						ForceNew:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 4096)),
					},
				},
			},
		},
		"tcp_reject": {
			Description: "Terminates a connection.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"active": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
						ForceNew: true,
					},
				},
			},
		},
		"set_forwarded_headers": {
			Description: "Adds 'X-Forwarded-For / -Proto / -Port' headers in your forwarded requests",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    100,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"active": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
						ForceNew: true,
					},
				},
			},
		},
	}
}

func loadBalancerActionsFromResourceData(d *schema.ResourceData) ([]upcloud.LoadBalancerAction, error) {
	a := make([]upcloud.LoadBalancerAction, 0)
	if _, ok := d.GetOk("actions.0"); !ok {
		return a, nil
	}

	// set_forwarded_headers action has to be iterated over first to avoid issues with actions ordering. This is because Managed Load Balancer evaluates actions in the same order
	// as they were set. But because each action has it's own, separate block in TF configuration, we cannot actually make sure they are ordered as the user intended.
	// This is not a big issue right now because all the actions except set_forwarded_headers are "final" (i.e. they end the chain and the next action is not evaluated).
	// So the only real use-case of having multiple actions is to have set_forwarded_headers action first, and then one of the "final" actions.
	// Therefore we work around the ordering problem by just making sure set_forwarded_headers actions are always set first.
	// TODO: Look for some more robust way of handling this when release a new major version
	for range d.Get("actions.0.set_forwarded_headers").([]interface{}) {
		a = append(a, request.NewLoadBalancerSetForwardedHeadersAction())
	}

	for _, v := range d.Get("actions.0.use_backend").([]interface{}) {
		v := v.(map[string]interface{})
		a = append(a, request.NewLoadBalancerUseBackendAction(v["backend_name"].(string)))
	}

	for _, v := range d.Get("actions.0.http_return").([]interface{}) {
		v := v.(map[string]interface{})
		a = append(a, request.NewLoadBalancerHTTPReturnAction(
			v["status"].(int),
			v["content_type"].(string),
			v["payload"].(string),
		))
	}

	for _, v := range d.Get("actions.0.http_redirect").([]interface{}) {
		v := v.(map[string]interface{})
		a = append(a, request.NewLoadBalancerHTTPRedirectAction(v["location"].(string)))
	}

	for range d.Get("actions.0.tcp_reject").([]interface{}) {
		a = append(a, request.NewLoadBalancerTCPRejectAction())
	}

	return a, nil
}

func setFrontendRuleActionsResourceData(d *schema.ResourceData, rule *upcloud.LoadBalancerFrontendRule) error {
	if len(rule.Actions) == 0 {
		return d.Set("actions", nil)
	}

	actions := make(map[string][]interface{})
	for _, a := range rule.Actions {
		t := string(a.Type)
		var v map[string]interface{}
		switch a.Type {
		case upcloud.LoadBalancerActionTypeUseBackend:
			v = map[string]interface{}{
				"backend_name": a.UseBackend.Backend,
			}
		case upcloud.LoadBalancerActionTypeHTTPRedirect:
			v = map[string]interface{}{
				"location": a.HTTPRedirect.Location,
			}
		case upcloud.LoadBalancerActionTypeHTTPReturn:
			v = map[string]interface{}{
				"content_type": a.HTTPReturn.ContentType,
				"status":       a.HTTPReturn.Status,
				"payload":      a.HTTPReturn.Payload,
			}
		case upcloud.LoadBalancerActionTypeTCPReject:
			v = map[string]interface{}{
				"active": true,
			}
		case upcloud.LoadBalancerActionTypeSetForwardedHeaders:
			v = map[string]interface{}{
				"active": true,
			}
		default:
			return fmt.Errorf("received unsupported action type '%s' %+v", a.Type, a)
		}
		if v != nil {
			actions[t] = append(actions[t], v)
		}
	}
	return d.Set("actions", []interface{}{actions})
}
