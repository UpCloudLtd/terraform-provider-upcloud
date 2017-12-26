package upcloud

import (
	"strconv"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudFirewallRuleCreate,
		Read:   resourceUpCloudFirewallRuleRead,
		Delete: resourceUpCloudFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"serverId": {
				Type:     schema.TypeString,
				Required: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"position": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"icmp_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_address_start": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_address_end": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_port_end": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_port_start": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_address_start": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_address_end": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_port_start": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_port_end": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUpCloudFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	createFirewallRuleRequest := &request.CreateFirewallRuleRequest{
		ServerUUID: d.Get("serverId").(string),
		FirewallRule: upcloud.FirewallRule{
			Direction: d.Get("direction").(string),
			Action:    d.Get("action").(string),
			Family:    d.Get("family").(string),
		},
	}

	if position, ok := d.GetOk("position"); ok {
		createFirewallRuleRequest.Position = position.(int)
	}

	if protocol, ok := d.GetOk("protocol"); ok {
		createFirewallRuleRequest.Protocol = protocol.(string)
	}

	if icmpType, ok := d.GetOk("icmp_type"); ok {
		createFirewallRuleRequest.ICMPType = icmpType.(string)
	}

	if destinationAddressStart, ok := d.GetOk("destination_address_start"); ok {
		createFirewallRuleRequest.DestinationAddressStart = destinationAddressStart.(string)
	}

	if destinationAddressEnd, ok := d.GetOk("destination_address_end"); ok {
		createFirewallRuleRequest.DestinationAddressEnd = destinationAddressEnd.(string)
	}

	if destinationPortStart, ok := d.GetOk("destination_port_start"); ok {
		createFirewallRuleRequest.DestinationPortStart = destinationPortStart.(string)
	}

	if destinationAddressStart, ok := d.GetOk("destination_port_end"); ok {
		createFirewallRuleRequest.DestinationAddressStart = destinationAddressStart.(string)
	}

	if sourceAddressStart, ok := d.GetOk("source_address_start"); ok {
		createFirewallRuleRequest.SourceAddressStart = sourceAddressStart.(string)
	}

	if sourceAddressEnd, ok := d.GetOk("source_address_end"); ok {
		createFirewallRuleRequest.SourceAddressEnd = sourceAddressEnd.(string)
	}

	if sourcePortStart, ok := d.GetOk("source_port_start"); ok {
		createFirewallRuleRequest.SourcePortStart = sourcePortStart.(string)
	}

	if sourcePortEnd, ok := d.GetOk("source_port_end"); ok {
		createFirewallRuleRequest.SourcePortEnd = sourcePortEnd.(string)
	}

	firewallRule, err := client.CreateFirewallRule(createFirewallRuleRequest)

	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(firewallRule.Position))

	return nil
}

func resourceUpCloudFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	position, err := strconv.Atoi(d.Id())

	if err != nil {
		return err
	}

	r := &request.GetFirewallRuleDetailsRequest{
		ServerUUID: d.Get("serverId").(string),
		Position:   position,
	}

	firewallRule, err := client.GetFirewallRuleDetails(r)

	if err != nil {
		return err
	}

	d.Set("action", firewallRule.Action)
	d.Set("comment", firewallRule.Comment)
	d.Set("destination_address_end", firewallRule.DestinationAddressEnd)
	d.Set("destination_address_start", firewallRule.DestinationAddressStart)
	d.Set("destination_port_end", firewallRule.DestinationPortEnd)
	d.Set("destination_port_start", firewallRule.DestinationPortStart)
	d.Set("direction", firewallRule.Direction)
	d.Set("family", firewallRule.Family)
	d.Set("icmp_type", firewallRule.ICMPType)
	d.Set("position", firewallRule.Position)
	d.Set("protocol", firewallRule.Protocol)
	d.Set("source_address_end", firewallRule.SourceAddressEnd)
	d.Set("source_address_start", firewallRule.SourceAddressStart)
	d.Set("source_port_end", firewallRule.SourcePortEnd)
	d.Set("source_port_start", firewallRule.SourcePortStart)

	return nil
}

func resourceUpCloudFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	position, err := strconv.Atoi(d.Id())

	if err != nil {
		return err
	}

	deleteFirewallRuleRequest := &request.DeleteFirewallRuleRequest{
		ServerUUID: d.Get("serverId").(string),
		Position:   position,
	}

	err = client.DeleteFirewallRule(deleteFirewallRuleRequest)

	if err != nil {
		return err
	}

	return nil
}
