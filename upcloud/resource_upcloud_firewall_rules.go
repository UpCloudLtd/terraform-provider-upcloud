package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUpCloudFirewallRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudFirewallRulesCreate,
		ReadContext:   resourceUpCloudFirewallRulesRead,
		UpdateContext: resourceUpCloudFirewallRulesUpdate,
		DeleteContext: resourceUpCloudFirewallRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Description: "The unique id of the server to be protected the firewall rules",
				Required:    true,
				ForceNew:    true,
			},
			"firewall_rule": {
				Type:     schema.TypeList,
				MaxItems: 1000,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": {
							Type:        schema.TypeString,
							Description: "The direction of network traffic this rule will be applied to",
							Required:    true,
							ForceNew:    true,
						},
						"action": {
							Type:        schema.TypeString,
							Description: "Action to take if the rule conditions are met",
							Required:    true,
							ForceNew:    true,
						},
						"family": {
							Type:        schema.TypeString,
							Description: "The address family of new firewall rule",
							Required:    true,
							ForceNew:    true,
						},
						"protocol": {
							Type:        schema.TypeString,
							Description: "The protocol this rule will be applied to",
							Optional:    true,
							ForceNew:    true,
							Default:     "tcp",
						},
						"icmp_type": {
							Type:        schema.TypeString,
							Description: "The ICMP type",
							Optional:    true,
							ForceNew:    true,
						},
						"source_address_start": {
							Type:        schema.TypeString,
							Description: "The source address range starts from this address",
							Optional:    true,
							ForceNew:    true,
						},
						"source_address_end": {
							Type:        schema.TypeString,
							Description: "The source address range ends from this address",
							Optional:    true,
							ForceNew:    true,
						},
						"source_port_end": {
							Type:        schema.TypeString,
							Description: "The source port range ends from this port number",
							Optional:    true,
							ForceNew:    true,
						},
						"source_port_start": {
							Type:        schema.TypeString,
							Description: "The source port range starts from this port number",
							Optional:    true,
							ForceNew:    true,
						},
						"destination_address_start": {
							Type:        schema.TypeString,
							Description: "The destination address range starts from this address",
							Optional:    true,
							ForceNew:    true,
						},
						"destination_address_end": {
							Type:        schema.TypeString,
							Description: "The destination address range ends from this address",
							Optional:    true,
							ForceNew:    true,
						},
						"destination_port_start": {
							Type:        schema.TypeString,
							Description: "The destination port range starts from this port number",
							Optional:    true,
							ForceNew:    true,
						},
						"destination_port_end": {
							Type:        schema.TypeString,
							Description: "The destination port range ends from this port number",
							Optional:    true,
							ForceNew:    true,
						},
						"comment": {
							Type:        schema.TypeString,
							Description: "Freeform comment string for the rule",
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceUpCloudFirewallRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	opts := &request.CreateFirewallRulesRequest{
		ServerUUID: d.Get("server_id").(string),
	}

	if v, ok := d.GetOk("firewall_rule"); ok {

		var firewallRules []upcloud.FirewallRule

		for _, frMap := range v.([]interface{}) {
			rule := frMap.(map[string]interface{})

			firewallRule := upcloud.FirewallRule{
				Action:                  rule["action"].(string),
				Comment:                 rule["comment"].(string),
				DestinationAddressStart: rule["destination_address_start"].(string),
				DestinationAddressEnd:   rule["destination_address_end"].(string),
				DestinationPortStart:    rule["destination_port_start"].(string),
				DestinationPortEnd:      rule["destination_port_end"].(string),
				Direction:               rule["direction"].(string),
				Family:                  rule["family"].(string),
				ICMPType:                rule["icmp_type"].(string),
				Protocol:                rule["protocol"].(string),
				SourceAddressStart:      rule["source_address_start"].(string),
				SourceAddressEnd:        rule["source_address_end"].(string),
				SourcePortStart:         rule["source_port_start"].(string),
				SourcePortEnd:           rule["source_port_end"].(string),
			}

			firewallRules = append(firewallRules, firewallRule)
		}

		opts.FirewallRules = firewallRules
	}

	err := client.CreateFirewallRules(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("server_id").(string))

	return resourceUpCloudFirewallRulesRead(ctx, d, meta)
}

func resourceUpCloudFirewallRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	opts := &request.GetFirewallRulesRequest{
		ServerUUID: d.Id(),
	}

	firewallRules, err := client.GetFirewallRules(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	var frMaps []map[string]interface{}

	for _, rule := range firewallRules.FirewallRules {

		frMap := map[string]interface{}{
			"action":                    rule.Action,
			"comment":                   rule.Comment,
			"destination_address_end":   rule.DestinationAddressEnd,
			"destination_address_start": rule.DestinationAddressStart,
			"destination_port_end":      rule.DestinationPortEnd,
			"destination_port_start":    rule.DestinationPortStart,
			"direction":                 rule.Direction,
			"family":                    rule.Family,
			"icmp_type":                 rule.ICMPType,
			"protocol":                  rule.Protocol,
			"source_address_end":        rule.SourceAddressEnd,
			"source_address_start":      rule.SourceAddressStart,
			"source_port_end":           rule.SourcePortEnd,
			"source_port_start":         rule.SourcePortStart,
		}

		frMaps = append(frMaps, frMap)
	}

	if err := d.Set("firewall_rule", frMaps); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUpCloudFirewallRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	opts := &request.CreateFirewallRulesRequest{
		ServerUUID: d.Id(),
	}

	err := client.CreateFirewallRules(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("firewall_rule") {

		v := d.Get("firewall_rule")
		var firewallRules []upcloud.FirewallRule

		for _, frMap := range v.([]interface{}) {
			rule := frMap.(map[string]interface{})

			firewallRule := upcloud.FirewallRule{
				Action:                  rule["action"].(string),
				Comment:                 rule["comment"].(string),
				DestinationAddressStart: rule["destination_address_start"].(string),
				DestinationAddressEnd:   rule["destination_address_end"].(string),
				DestinationPortStart:    rule["destination_port_start"].(string),
				DestinationPortEnd:      rule["destination_port_end"].(string),
				Direction:               rule["direction"].(string),
				Family:                  rule["family"].(string),
				ICMPType:                rule["icmp_type"].(string),
				Protocol:                rule["protocol"].(string),
				SourceAddressStart:      rule["source_address_start"].(string),
				SourceAddressEnd:        rule["source_address_end"].(string),
				SourcePortStart:         rule["source_port_start"].(string),
				SourcePortEnd:           rule["source_port_end"].(string),
			}

			firewallRules = append(firewallRules, firewallRule)
		}

		opts.FirewallRules = firewallRules
	}

	err = client.CreateFirewallRules(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudFirewallRulesRead(ctx, d, meta)
}

func resourceUpCloudFirewallRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	opts := &request.CreateFirewallRulesRequest{
		ServerUUID:    d.Id(),
		FirewallRules: nil,
	}

	err := client.CreateFirewallRules(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
