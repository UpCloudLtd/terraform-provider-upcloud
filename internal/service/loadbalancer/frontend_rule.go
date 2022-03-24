package loadbalancer

import (
	"context"
	"log"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceFrontendRule() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer frontend rule",
		CreateContext: resourceFrontendRuleCreate,
		ReadContext:   resourceFrontendRuleRead,
		UpdateContext: resourceFrontendRuleUpdate,
		DeleteContext: resourceFrontendRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer": {
				Description: "ID of the load balancer to which the frontend is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"frontend_name": {
				Description: "Name of the load balancer frontend to which the rule is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the frontend must be unique within the load balancer service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"priority": {
				Description:      "Rule with the higher priority goes first. Rules with the same priority processed in alphabetical order.",
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
			},
			"matchers": {
				Description: "Set of rule matchers. if rule doesn't have matchers, then action applies to all incoming requests.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: frontendRuleMatchersSchema(),
				},
			},
			"actions": {
				Description: "Set of rule actions.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: frontendRuleActionsSchema(),
				},
			},
		},
	}
}

func resourceFrontendRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	matchers, err := loadBalancerMatchersFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	actions, err := loadBalancerActionsFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	rule, err := svc.CreateLoadBalancerFrontendRule(&request.CreateLoadBalancerFrontendRuleRequest{
		ServiceUUID:  d.Get("loadbalancer").(string),
		FrontendName: d.Get("frontend_name").(string),
		Rule: request.LoadBalancerFrontendRule{
			Name:     d.Get("name").(string),
			Priority: d.Get("priority").(int),
			Matchers: matchers,
			Actions:  actions,
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	d.SetId(rule.Name)

	log.Printf("[INFO] frontend rule '%s' created", rule.Name)
	return diags
}

func resourceFrontendRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	rule, err := svc.GetLoadBalancerFrontendRule(&request.GetLoadBalancerFrontendRuleRequest{
		ServiceUUID:  d.Get("loadbalancer").(string),
		FrontendName: d.Get("frontend_name").(string),
		Name:         d.Id(),
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	d.SetId(rule.Name)

	return diags
}

func resourceFrontendRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	// name and priority fields doesn't force replacement and can be updated in-place
	rule, err := svc.ModifyLoadBalancerFrontendRule(&request.ModifyLoadBalancerFrontendRuleRequest{
		ServiceUUID:  d.Get("loadbalancer").(string),
		FrontendName: d.Get("frontend_name").(string),
		Name:         d.Id(),
		Rule: request.ModifyLoadBalancerFrontendRule{
			Name:     d.Get("name").(string),
			Priority: d.Get("priority").(int),
		},
	},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.Name)

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] frontend rule '%s' updated", rule.Name)
	return diags
}

func resourceFrontendRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	log.Printf("[INFO] deleting frontend rule '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerFrontendRule(&request.DeleteLoadBalancerFrontendRuleRequest{
		ServiceUUID:  d.Get("loadbalancer").(string),
		FrontendName: d.Get("frontend_name").(string),
		Name:         d.Id(),
	}))
}

func setFrontendRuleResourceData(d *schema.ResourceData, rule *upcloud.LoadBalancerFrontendRule) (diags diag.Diagnostics) {
	if err := d.Set("name", rule.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("priority", rule.Priority); err != nil {
		return diag.FromErr(err)
	}

	if err := setFrontendRuleMatchersResourceData(d, rule); err != nil {
		return diag.FromErr(err)
	}

	if err := setFrontendRuleActionsResourceData(d, rule); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
