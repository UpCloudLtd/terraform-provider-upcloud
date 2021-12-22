package upcloud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudFirewallRules() *schema.Resource {
	return &schema.Resource{
		Description: `This resource represents a generated list of UpCloud firewall rules. 
		Firewall rules are used in conjunction with UpCloud servers. 
		Each server has its own firewall rules. 
		The firewall is enabled on all network interfaces except ones attached to private virtual networks. 
		The maximum number of firewall rules per server is 1000.`,
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
							ValidateFunc: validation.StringInSlice([]string{"", "tcp", "udp", "icmp"}, false),
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
						"source_port_start": {
							Type:             schema.TypeString,
							Description:      "The source port range starts from this port number",
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: firewallRuleValidateOptionalPort,
						},
						"source_port_end": {
							Type:             schema.TypeString,
							Description:      "The source port range ends from this port number",
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: firewallRuleValidateOptionalPort,
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
							Type:             schema.TypeString,
							Description:      "The destination port range starts from this port number",
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: firewallRuleValidateOptionalPort,
						},
						"destination_port_end": {
							Type:             schema.TypeString,
							Description:      "The destination port range ends from this port number",
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: firewallRuleValidateOptionalPort,
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
			"destination_port_start":    rule.DestinationPortStart,
			"destination_port_end":      rule.DestinationPortEnd,
			"direction":                 rule.Direction,
			"family":                    rule.Family,
			"icmp_type":                 rule.ICMPType,
			"protocol":                  rule.Protocol,
			"source_address_end":        rule.SourceAddressEnd,
			"source_address_start":      rule.SourceAddressStart,
			"source_port_start":         rule.SourcePortStart,
			"source_port_end":           rule.SourcePortEnd,
		}

		frMaps = append(frMaps, frMap)
	}

	if err := d.Set("firewall_rule", frMaps); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("server_id", d.Id()); err != nil {
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

func firewallRuleValidateOptionalPort(v interface{}, path cty.Path) diag.Diagnostics {
	const (
		portMin int = 1
		portMax int = 65535
	)
	var diags diag.Diagnostics
	val, ok := v.(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Bad type",
			Detail:        "expected type to be string",
			AttributePath: path,
		})
		return diags
	}

	if val == "" {
		return diags
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Bad port format",
			Detail:        fmt.Sprintf("%s is not valid number", val),
			AttributePath: path,
		})
		return diags
	}

	if portMin > i || i > portMax {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Bad port",
			Detail:        fmt.Sprintf("%s is not within valid port range %d - %d", val, portMin, portMax),
			AttributePath: path,
		})
		return diags
	}

	return diags
}
