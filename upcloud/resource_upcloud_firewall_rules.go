package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"time"
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
							Type:         schema.TypeString,
							Description:  "The direction of network traffic this rule will be applied to",
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"in", "out"}, false),
						},
						"action": {
							Type:         schema.TypeString,
							Description:  "Action to take if the rule conditions are met",
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"accept", "drop"}, false),
						},
						"family": {
							Type:         schema.TypeString,
							Description:  "The address family of new firewall rule",
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, false),
						},
						"protocol": {
							Type:         schema.TypeString,
							Description:  "The protocol this rule will be applied to",
							Optional:     true,
							ForceNew:     true,
							Default:      "tcp",
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "icmp"}, false),
						},
						"icmp_type": {
							Type:         schema.TypeString,
							Description:  "The ICMP type",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"source_address_start": {
							Type:         schema.TypeString,
							Description:  "The source address range starts from this address",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty),
						},
						"source_address_end": {
							Type:         schema.TypeString,
							Description:  "The source address range ends from this address",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty),
						},
						"source_port_end": {
							Type:         schema.TypeInt,
							Description:  "The source port range ends from this port number",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"source_port_start": {
							Type:         schema.TypeInt,
							Description:  "The source port range starts from this port number",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"destination_address_start": {
							Type:         schema.TypeString,
							Description:  "The destination address range starts from this address",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty),
						},
						"destination_address_end": {
							Type:         schema.TypeString,
							Description:  "The destination address range ends from this address",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.Any(validation.IsIPv4Address, validation.IsIPv6Address, validation.StringIsEmpty),
						},
						"destination_port_start": {
							Type:         schema.TypeInt,
							Description:  "The destination port range starts from this port number",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"destination_port_end": {
							Type:         schema.TypeInt,
							Description:  "The destination port range ends from this port number",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"comment": {
							Type:         schema.TypeString,
							Description:  "Freeform comment string for the rule",
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 250),
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

			destinationPortStart := strconv.Itoa(rule["destination_port_start"].(int))
			if destinationPortStart == "0" {
				destinationPortStart = ""
			}

			destinationPortEnd := strconv.Itoa(rule["destination_port_end"].(int))
			if destinationPortEnd == "0" {
				destinationPortEnd = ""
			}

			sourcePortStart := strconv.Itoa(rule["source_port_start"].(int))
			if sourcePortStart == "0" {
				sourcePortStart = ""
			}

			sourcePortEnd := strconv.Itoa(rule["source_port_end"].(int))
			if sourcePortEnd == "0" {
				sourcePortEnd = ""
			}

			firewallRule := upcloud.FirewallRule{
				Action:                  rule["action"].(string),
				Comment:                 rule["comment"].(string),
				DestinationAddressStart: rule["destination_address_start"].(string),
				DestinationAddressEnd:   rule["destination_address_end"].(string),
				DestinationPortStart:    destinationPortStart,
				DestinationPortEnd:      destinationPortEnd,
				Direction:               rule["direction"].(string),
				Family:                  rule["family"].(string),
				ICMPType:                rule["icmp_type"].(string),
				Protocol:                rule["protocol"].(string),
				SourceAddressStart:      rule["source_address_start"].(string),
				SourceAddressEnd:        rule["source_address_end"].(string),
				SourcePortStart:         sourcePortStart,
				SourcePortEnd:           sourcePortEnd,
			}

			firewallRules = append(firewallRules, firewallRule)
		}

		opts.FirewallRules = firewallRules
	}

	if _, err := client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:           opts.ServerUUID,
		UndesiredState: upcloud.ServerStateMaintenance,
		Timeout:        time.Minute * 5,
	}); err != nil {
		return diag.FromErr(err)
	}

	if err := client.CreateFirewallRules(opts); err != nil {
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
			"direction":                 rule.Direction,
			"family":                    rule.Family,
			"icmp_type":                 rule.ICMPType,
			"protocol":                  rule.Protocol,
			"source_address_end":        rule.SourceAddressEnd,
			"source_address_start":      rule.SourceAddressStart,
		}

		if rule.DestinationPortEnd != "" {
			value, err := strconv.Atoi(rule.DestinationPortEnd)
			if err != nil {
				diag.FromErr(err)
			}
			frMap["destination_port_end"] = value
		}

		if rule.DestinationPortStart != "" {
			value, err := strconv.Atoi(rule.DestinationPortStart)
			if err != nil {
				diag.FromErr(err)
			}
			frMap["destination_port_start"] = value
		}

		if rule.SourcePortEnd != "" {
			value, err := strconv.Atoi(rule.SourcePortEnd)
			if err != nil {
				diag.FromErr(err)
			}
			frMap["source_port_end"] = value
		}

		if rule.SourcePortStart != "" {
			value, err := strconv.Atoi(rule.SourcePortStart)
			if err != nil {
				diag.FromErr(err)
			}
			frMap["source_port_start"] = value
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

			destinationPortStart := strconv.Itoa(rule["destination_port_start"].(int))
			if destinationPortStart == "0" {
				destinationPortStart = ""
			}

			destinationPortEnd := strconv.Itoa(rule["destination_port_end"].(int))
			if destinationPortEnd == "0" {
				destinationPortEnd = ""
			}

			sourcePortStart := strconv.Itoa(rule["source_port_start"].(int))
			if sourcePortStart == "0" {
				sourcePortStart = ""
			}

			sourcePortEnd := strconv.Itoa(rule["source_port_end"].(int))
			if sourcePortEnd == "0" {
				sourcePortEnd = ""
			}

			firewallRule := upcloud.FirewallRule{
				Action:                  rule["action"].(string),
				Comment:                 rule["comment"].(string),
				DestinationAddressStart: rule["destination_address_start"].(string),
				DestinationAddressEnd:   rule["destination_address_end"].(string),
				DestinationPortStart:    destinationPortStart,
				DestinationPortEnd:      destinationPortEnd,
				Direction:               rule["direction"].(string),
				Family:                  rule["family"].(string),
				ICMPType:                rule["icmp_type"].(string),
				Protocol:                rule["protocol"].(string),
				SourceAddressStart:      rule["source_address_start"].(string),
				SourceAddressEnd:        rule["source_address_end"].(string),
				SourcePortStart:         sourcePortStart,
				SourcePortEnd:           sourcePortEnd,
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

	if _, err := client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:           opts.ServerUUID,
		UndesiredState: upcloud.ServerStateMaintenance,
		Timeout:        time.Minute * 5,
	}); err != nil {
		return diag.FromErr(err)
	}

	if err := client.CreateFirewallRules(opts); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
