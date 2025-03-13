package upcloud

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/firewall"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/gateway"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/managedobjectstorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/network"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/objectstorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/tag"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_USERNAME", nil),
				Description: usernameDescription,
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_PASSWORD", nil),
				Description: passwordDescription,
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UPCLOUD_TOKEN", nil),
				Description: tokenDescription,
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
			"request_timeout_sec": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     120,
				Description: "The duration (in seconds) that the provider waits for an HTTP request towards UpCloud API to complete. Defaults to 120 seconds",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"upcloud_firewall_rules":                         firewall.ResourceFirewallRules(),
			"upcloud_tag":                                    tag.ResourceTag(),
			"upcloud_gateway":                                gateway.ResourceGateway(),
			"upcloud_gateway_connection":                     gateway.ResourceConnection(),
			"upcloud_gateway_connection_tunnel":              gateway.ResourceTunnel(),
			"upcloud_object_storage":                         objectstorage.ResourceObjectStorage(),
			"upcloud_managed_database_postgresql":            database.ResourcePostgreSQL(),
			"upcloud_managed_database_mysql":                 database.ResourceMySQL(),
			"upcloud_managed_database_redis":                 database.ResourceRedis(),
			"upcloud_managed_database_opensearch":            database.ResourceOpenSearch(),
			"upcloud_managed_database_valkey":                database.ResourceValkey(),
			"upcloud_managed_database_user":                  database.ResourceUser(),
			"upcloud_managed_database_logical_database":      database.ResourceLogicalDatabase(),
			"upcloud_managed_object_storage":                 managedobjectstorage.ResourceManagedObjectStorage(),
			"upcloud_managed_object_storage_user":            managedobjectstorage.ResourceManagedObjectStorageUser(),
			"upcloud_managed_object_storage_user_access_key": managedobjectstorage.ResourceManagedObjectStorageUserAccessKey(),
			"upcloud_managed_object_storage_user_policy":     managedobjectstorage.ResourceManagedObjectStorageUserPolicy(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"upcloud_networks": network.DataSourceNetworks(),
			"upcloud_tags":     tag.DataSourceTags(),
			"upcloud_managed_database_opensearch_indices":  database.DataSourceOpenSearchIndices(),
			"upcloud_managed_database_mysql_sessions":      database.DataSourceSessionsMySQL(),
			"upcloud_managed_database_postgresql_sessions": database.DataSourceSessionsPostgreSQL(),
			"upcloud_managed_database_redis_sessions":      database.DataSourceSessionsRedis(),
			"upcloud_managed_database_valkey_sessions":     database.DataSourceSessionsValkey(),
			"upcloud_managed_object_storage_policies":      managedobjectstorage.DataSourceManagedObjectStoragePolicies(),
		},

		ConfigureContextFunc: providerConfigureWithDefaultUserAgent,
	}
}

func providerConfigureWithDefaultUserAgent(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return ProviderConfigure(ctx, d, config.DefaultUserAgent())
}

func ProviderConfigure(_ context.Context, d *schema.ResourceData, userAgents ...string) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	requestTimeout := time.Duration(d.Get("request_timeout_sec").(int)) * time.Second

	cfg := config.Config{
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
		Token:    d.Get("token").(string),
	}

	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = time.Duration(d.Get("retry_wait_min_sec").(int)) * time.Second
	httpClient.RetryWaitMax = time.Duration(d.Get("retry_wait_max_sec").(int)) * time.Second
	httpClient.RetryMax = d.Get("retry_max").(int)

	svc, err := cfg.NewUpCloudServiceConnection(
		httpClient.HTTPClient,
		requestTimeout,
		userAgents...,
	)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return svc, diags
}
