package upcloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	upcloudAPITimeout = time.Second * 120
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_USERNAME", nil),
				Description: "UpCloud username with API access",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_PASSWORD", nil),
				Description: "Password for UpCloud API user",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"upcloud_server":         resourceUpCloudServer(),
			"upcloud_router":         resourceUpCloudRouter(),
			"upcloud_storage":        resourceUpCloudStorage(),
			"upcloud_firewall_rules": resourceUpCloudFirewallRules(),
			//"upcloud_tag":           resourceUpCloudTag(),
			"upcloud_network":             resourceUpCloudNetwork(),
			"upcloud_floating_ip_address": resourceUpCloudFloatingIPAddress(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"upcloud_zone":         dataSourceUpCloudZone(),
			"upcloud_zones":        dataSourceUpCloudZones(),
			"upcloud_networks":     dataSourceNetworks(),
			"upcloud_hosts":        dataSourceUpCloudHosts(),
			"upcloud_ip_addresses": dataSourceUpCloudIPAddresses(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	var diags diag.Diagnostics

	config := Config{
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}

	client := client.New(d.Get("username").(string), d.Get("password").(string))
	client.SetTimeout(upcloudAPITimeout)

	service := service.New(client)

	_, err := config.checkLogin(service)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return service, diags
}
