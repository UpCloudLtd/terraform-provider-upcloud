package upcloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

const (
	upcloudAPITimeout = time.Second * 120
	version           = "0.1.0"
)

type userAgentTransport struct{}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", fmt.Sprintf("terraform-provider-upcloud/%s", version))
	return cleanhttp.DefaultTransport().RoundTrip(req)
}

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
			"retry_wait_min_sec": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Minimum time to wait between retries",
			},
			"retry_wait_max_sec": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Maximum time to wait between retries",
			},
			"retry_max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     4,
				Description: "Maximum number of retries",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"upcloud_server":              resourceUpCloudServer(),
			"upcloud_router":              resourceUpCloudRouter(),
			"upcloud_storage":             resourceUpCloudStorage(),
			"upcloud_firewall_rules":      resourceUpCloudFirewallRules(),
			"upcloud_tag":                 resourceUpCloudTag(),
			"upcloud_network":             resourceUpCloudNetwork(),
			"upcloud_floating_ip_address": resourceUpCloudFloatingIPAddress(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"upcloud_zone":         dataSourceUpCloudZone(),
			"upcloud_zones":        dataSourceUpCloudZones(),
			"upcloud_networks":     dataSourceNetworks(),
			"upcloud_hosts":        dataSourceUpCloudHosts(),
			"upcloud_ip_addresses": dataSourceUpCloudIPAddresses(),
			"upcloud_tags":         dataSourceUpCloudTags(),
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

	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = time.Duration(d.Get("retry_wait_min_sec").(int)) * time.Second
	httpClient.RetryWaitMax = time.Duration(d.Get("retry_wait_max_sec").(int)) * time.Second
	httpClient.RetryMax = d.Get("retry_max").(int)
	httpClient.HTTPClient = &http.Client{Transport: &userAgentTransport{}}

	client := client.NewWithHTTPClient(d.Get("username").(string), d.Get("password").(string), httpClient.HTTPClient)
	client.SetTimeout(upcloudAPITimeout)

	service := service.New(client)

	_, err := config.checkLogin(service)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return service, diags
}
