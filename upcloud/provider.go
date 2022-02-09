package upcloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

const (
	upcloudAPITimeout                       time.Duration = time.Second * 120
	upcloudServerNotFoundErrorCode          string        = "SERVER_NOT_FOUND"
	upcloudStorageNotFoundErrorCode         string        = "STORAGE_NOT_FOUND"
	upcloudNetworkNotFoundErrorCode         string        = "NETWORK_NOT_FOUND"
	upcloudRouterNotFoundErrorCode          string        = "ROUTER_NOT_FOUND"
	upcloudObjectStorageNotFoundErrorCode   string        = "OBJECT_STORAGE_NOT_FOUND"
	upcloudDatabaseNotFoundErrorCode        string        = "SERVICE_NOT_FOUND"
	upcloudLogicalDatabaseNotFoundErrorCode string        = "DB_NOT_FOUND"
	upcloudDatabaseUserNotFoundErrorCode    string        = "USER_NOT_FOUND"
	upcloudIPAddressNotFoundErrorCode       string        = "IP_ADDRESS_NOT_FOUND"
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
			"upcloud_server":                            resourceUpCloudServer(),
			"upcloud_router":                            resourceUpCloudRouter(),
			"upcloud_storage":                           resourceUpCloudStorage(),
			"upcloud_firewall_rules":                    resourceUpCloudFirewallRules(),
			"upcloud_tag":                               resourceUpCloudTag(),
			"upcloud_network":                           resourceUpCloudNetwork(),
			"upcloud_floating_ip_address":               resourceUpCloudFloatingIPAddress(),
			"upcloud_object_storage":                    resourceUpCloudObjectStorage(),
			"upcloud_managed_database_postgresql":       resourceUpCloudManagedDatabasePostgreSQL(),
			"upcloud_managed_database_mysql":            resourceUpCloudManagedDatabaseMySQL(),
			"upcloud_managed_database_user":             resourceUpCloudManagedDatabaseUser(),
			"upcloud_managed_database_logical_database": resourceUpCloudManagedDatabaseLogicalDatabase(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"upcloud_zone":         dataSourceUpCloudZone(),
			"upcloud_zones":        dataSourceUpCloudZones(),
			"upcloud_networks":     dataSourceNetworks(),
			"upcloud_hosts":        dataSourceUpCloudHosts(),
			"upcloud_ip_addresses": dataSourceUpCloudIPAddresses(),
			"upcloud_tags":         dataSourceUpCloudTags(),
			"upcloud_storage":      dataSourceUpCloudStorage(),
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

func newUpCloudServiceConnection(username, password string, httpClient *http.Client) *service.Service {
	client := client.NewWithHTTPClient(username, password, httpClient)
	client.UserAgent = fmt.Sprintf("terraform-provider-upcloud/%s", config.Version)
	client.SetTimeout(upcloudAPITimeout)

	return service.New(client)
}

func isProviderAccountSubaccount(s *service.Service) (bool, error) {
	account, err := s.GetAccount()
	if err != nil {
		return false, err
	}
	a, err := s.GetAccountDetails(&request.GetAccountDetailsRequest{Username: account.UserName})
	if err != nil {
		return false, err
	}
	return a.IsSubaccount(), nil
}

func diagWarningFromUpcloudErr(err *upcloud.Error, details string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  err.ErrorMessage,
		Detail:   details,
	}
}

func diagBindingRemovedWarningFromUpcloudErr(err *upcloud.Error, name string) diag.Diagnostic {
	return diagWarningFromUpcloudErr(err,
		fmt.Sprintf("Binding to an existing remote object '%s' will be removed from the state. Next plan will include action to re-create the object if you choose to keep it in config.", name),
	)
}
