package upcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudFirewallRuleCreate,
		Read:   resourceUpCloudFirewallRuleRead,
		Update: resourceUpCloudFirewallRuleUpdate,
		Delete: resourceUpCloudFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"position": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_address_start": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_address_end": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_port_end": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_port_start": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_address_start": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_address_end": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_port_start": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_port_end": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"icmp_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUpCloudFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
