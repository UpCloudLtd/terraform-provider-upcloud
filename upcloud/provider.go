package upcloud

import (
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	upcloudAPITimeout = time.Second * 240
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
				Description: "Password for Upcloud API user",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"upcloud_server":        resourceUpCloudServer(),
			"upcloud_storage":       resourceUpCloudStorage(),
			"upcloud_firewall_rule": resourceUpCloudFirewallRule(),
			"upcloud_plan":          resourceUpCloudPlan(),
			"upcloud_price":         resourceUpCloudPrice(),
			"upcloud_price_zone":    resourceUpCloudPriceZone(),
			"upcloud_tag":           resourceUpCloudTag(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}

	client := client.New(d.Get("username").(string), d.Get("password").(string))
	client.SetTimeout(upcloudAPITimeout)

	service := service.New(client)

	_, err := config.checkLogin(service)

	return service, err
}
