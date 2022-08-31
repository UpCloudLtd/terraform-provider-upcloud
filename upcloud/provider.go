package upcloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/cloud"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/firewall"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/ip"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/loadbalancer"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/network"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/objectstorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/router"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/server"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/storage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/tag"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

const upcloudAPITimeout time.Duration = time.Second * 120

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_USERNAME", nil),
				Description: "UpCloud username with API access. Can also be configured using the `UPCLOUD_USERNAME` environment variable.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_PASSWORD", nil),
				Description: "Password for UpCloud API user. Can also be configured using the `UPCLOUD_PASSWORD` environment variable.",
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
			"upcloud_server":                                  server.ResourceServer(),
			"upcloud_router":                                  router.ResourceRouter(),
			"upcloud_storage":                                 storage.ResourceStorage(),
			"upcloud_firewall_rules":                          firewall.ResourceFirewallRules(),
			"upcloud_tag":                                     tag.ResourceTag(),
			"upcloud_network":                                 network.ResourceNetwork(),
			"upcloud_floating_ip_address":                     ip.ResourceFloatingIPAddress(),
			"upcloud_object_storage":                          objectstorage.ResourceObjectStorage(),
			"upcloud_managed_database_postgresql":             database.ResourcePostgreSQL(),
			"upcloud_managed_database_mysql":                  database.ResourceMySQL(),
			"upcloud_managed_database_user":                   database.ResourceUser(),
			"upcloud_managed_database_logical_database":       database.ResourceLogicalDatabase(),
			"upcloud_loadbalancer":                            loadbalancer.ResourceLoadBalancer(),
			"upcloud_loadbalancer_resolver":                   loadbalancer.ResourceResolver(),
			"upcloud_loadbalancer_backend":                    loadbalancer.ResourceBackend(),
			"upcloud_loadbalancer_static_backend_member":      loadbalancer.ResourceStaticBackendMember(),
			"upcloud_loadbalancer_dynamic_backend_member":     loadbalancer.ResourceDynamicBackendMember(),
			"upcloud_loadbalancer_frontend":                   loadbalancer.ResourceFrontend(),
			"upcloud_loadbalancer_frontend_rule":              loadbalancer.ResourceFrontendRule(),
			"upcloud_loadbalancer_frontend_tls_config":        loadbalancer.ResourceFrontendTLSConfig(),
			"upcloud_loadbalancer_dynamic_certificate_bundle": loadbalancer.ResourceDynamicCertificateBundle(),
			"upcloud_loadbalancer_manual_certificate_bundle":  loadbalancer.ResourceManualCertificateBundle(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"upcloud_zone":         cloud.DataSourceZone(),
			"upcloud_zones":        cloud.DataSourceZones(),
			"upcloud_networks":     network.DataSourceNetworks(),
			"upcloud_hosts":        cloud.DataSourceHosts(),
			"upcloud_ip_addresses": ip.DataSourceIPAddresses(),
			"upcloud_tags":         tag.DataSourceTags(),
			"upcloud_storage":      storage.DataSourceStorage(),
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

	service := newUpCloudServiceConnection(
		d.Get("username").(string),
		d.Get("password").(string),
		httpClient.HTTPClient,
	)

	_, err := config.checkLogin(service)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return service, diags
}

func newUpCloudServiceConnection(username, password string, httpClient *http.Client) *service.ServiceContext {
	client := client.NewWithHTTPClientContext(username, password, httpClient)
	client.UserAgent = fmt.Sprintf("terraform-provider-upcloud/%s", config.Version)
	client.SetTimeout(upcloudAPITimeout)

	return service.NewWithContext(client)
}
