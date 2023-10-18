package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		CustomizeDiff: customdiff.All(
			// Validate http_redirect fields here, because ExactlyOneOf does not work when MaxItems > 1
			validateHTTPRedirectChange,
			validateActionsNotEmpty,
		),
		Schema: map[string]*schema.Schema{
			"frontend": {
				Description: "ID of the load balancer frontend to which the rule is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the frontend rule must be unique within the load balancer service.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNameDiagFunc,
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

	var serviceID, feName string
	if err := utils.UnmarshalID(d.Get("frontend").(string), &serviceID, &feName); err != nil {
		return diag.FromErr(err)
	}

	rule, err := svc.CreateLoadBalancerFrontendRule(ctx, &request.CreateLoadBalancerFrontendRuleRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
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

	d.SetId(utils.MarshalID(serviceID, feName, rule.Name))

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend rule created", map[string]interface{}{"name": rule.Name, "service_uuid": serviceID, "fe_name": feName})
	return diags
}

func resourceFrontendRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	rule, err := svc.GetLoadBalancerFrontendRule(ctx, &request.GetLoadBalancerFrontendRuleRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(utils.MarshalID(serviceID, feName, rule.Name))

	if err = d.Set("frontend", utils.MarshalID(serviceID, feName)); err != nil {
		return diag.FromErr(err)
	}

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceFrontendRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}
	// name and priority fields doesn't force replacement and can be updated in-place
	rule, err := svc.ModifyLoadBalancerFrontendRule(ctx, &request.ModifyLoadBalancerFrontendRuleRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
		Rule: request.ModifyLoadBalancerFrontendRule{
			Name:     d.Get("name").(string),
			Priority: upcloud.IntPtr(d.Get("priority").(int)),
		},
	},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, feName, rule.Name))

	if diags = setFrontendRuleResourceData(d, rule); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "frontend rule updated", map[string]interface{}{"name": rule.Name, "service_uuid": serviceID, "fe_name": feName})
	return diags
}

func resourceFrontendRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, feName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &feName, &name); err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "deleting frontend rule", map[string]interface{}{"name": name, "service_uuid": serviceID, "fe_name": feName})

	return diag.FromErr(svc.DeleteLoadBalancerFrontendRule(ctx, &request.DeleteLoadBalancerFrontendRuleRequest{
		ServiceUUID:  serviceID,
		FrontendName: feName,
		Name:         name,
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
