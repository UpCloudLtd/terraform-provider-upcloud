package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/cloud"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/filestorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/firewall"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/ip"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/kubernetes"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/loadbalancer"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/managedobjectstorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/network"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/networkpeering"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/router"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/server"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/servergroup"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/storage"
	"github.com/UpCloudLtd/upcloud-go-api/credentials"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	usernameDescription       = "UpCloud username with API access. Can also be configured using the `UPCLOUD_USERNAME` environment variable."
	passwordDescription       = "Password for UpCloud API user. Can also be configured using the `UPCLOUD_PASSWORD` environment variable."
	tokenDescription          = "Token for authenticating to UpCloud API. Can also be configured using the `UPCLOUD_TOKEN` environment variable or using the system keyring. Use `upctl account login` command to save a token to the system keyring. (EXPERIMENTAL)"
	requestTimeoutDescription = "The duration (in seconds) that the provider waits for an HTTP request towards UpCloud API to complete. Defaults to 120 seconds"
)

type upcloudProviderModel struct {
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	Token             types.String `tfsdk:"token"`
	RetryWaitMinSec   types.Int64  `tfsdk:"retry_wait_min_sec"`
	RetryWaitMaxSec   types.Int64  `tfsdk:"retry_wait_max_sec"`
	RetryMax          types.Int64  `tfsdk:"retry_max"`
	RequestTimeoutSec types.Int64  `tfsdk:"request_timeout_sec"`
}

type upcloudProvider struct {
	userAgent string
}

var (
	_ provider.Provider                       = &upcloudProvider{}
	_ provider.ProviderWithEphemeralResources = &upcloudProvider{}
)

func New() provider.Provider {
	return &upcloudProvider{
		userAgent: config.DefaultUserAgent(),
	}
}

func NewWithUserAgent(userAgent string) provider.Provider {
	return &upcloudProvider{
		userAgent: userAgent,
	}
}

func (p *upcloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "upcloud"
	resp.Version = config.Version
}

func (p *upcloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: usernameDescription,
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: passwordDescription,
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: tokenDescription,
				Optional:    true,
			},
			"retry_wait_min_sec": schema.Int64Attribute{
				Optional:    true,
				Description: "Minimum time to wait between retries",
			},
			"retry_wait_max_sec": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum time to wait between retries",
			},
			"retry_max": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of retries",
			},
			"request_timeout_sec": schema.Int64Attribute{
				Optional:    true,
				Description: requestTimeoutDescription,
			},
		},
	}
}

func withInt64Default(val types.Int64, def int64) int64 {
	if val.IsNull() {
		return def
	}
	return val.ValueInt64()
}

func (p *upcloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model upcloudProviderModel
	if diags := req.Config.Get(ctx, &model); diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	requestTimeout := time.Duration(withInt64Default(model.RequestTimeoutSec, 120)) * time.Second
	creds, err := credentials.Parse(credentials.Credentials{
		Username: model.Username.ValueString(),
		Password: model.Password.ValueString(),
		Token:    model.Token.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("No credentials defined", err.Error())
		return
	}
	cfg := config.NewFromCredentials(creds)

	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = time.Duration(withInt64Default(model.RetryWaitMinSec, 1)) * time.Second
	httpClient.RetryWaitMax = time.Duration(withInt64Default(model.RetryWaitMaxSec, 30)) * time.Second
	httpClient.RetryMax = int(withInt64Default(model.RetryMax, 4))

	service, err := cfg.NewUpCloudServiceConnection(
		httpClient.HTTPClient,
		requestTimeout,
		p.userAgent,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to UpCloud API with given credentials", err.Error())
	}

	tflog.Info(ctx, "UpCloud service connection configured for plugin framework provider", map[string]interface{}{"http_client": fmt.Sprintf("%#v", httpClient), "request_timeout": requestTimeout})

	resp.ResourceData = service
	resp.DataSourceData = service
}

func (p *upcloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		database.NewMySQLResource,
		database.NewOpenSearchResource,
		database.NewPostgresResource,
		database.NewValkeyResource,
		firewall.NewFirewallRulesResource,
		ip.NewFloatingIPAddressResource,
		kubernetes.NewKubernetesClusterResource,
		kubernetes.NewKubernetesNodeGroupResource,
		loadbalancer.NewBackendTLSConfigResource,
		loadbalancer.NewBackendDynamicMemberResource,
		loadbalancer.NewBackendStaticMemberResource,
		loadbalancer.NewDynamicCertificateBundleResource,
		loadbalancer.NewBackendResource,
		loadbalancer.NewFrontendResource,
		loadbalancer.NewFrontendRuleResource,
		loadbalancer.NewFrontendTLSConfigResource,
		loadbalancer.NewLoadBalancerResource,
		loadbalancer.NewManualCertificateBundleResource,
		loadbalancer.NewResolverResource,
		managedobjectstorage.NewManagedObjectStorageResource,
		managedobjectstorage.NewBucketResource,
		managedobjectstorage.NewCustomDomainResource,
		managedobjectstorage.NewPolicyResource,
		managedobjectstorage.NewUserResource,
		managedobjectstorage.NewUserAccessKeyResource,
		managedobjectstorage.NewUserPolicyResource,
		network.NewNetworkResource,
		networkpeering.NewNetworkPeeringResource,
		router.NewRouterResource,
		server.NewServerResource,
		servergroup.NewServerGroupResource,
		storage.NewStorageResource,
		storage.NewStorageTemplateResource,
		storage.NewStorageBackupResource,
		filestorage.NewFileStorageResource,
		filestorage.NewFileStorageShareResource,
		filestorage.NewFileStorageShareACLResource,
	}
}

func (p *upcloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloud.NewHostsDataSource,
		cloud.NewZoneDataSource,
		cloud.NewZonesDataSource,
		ip.NewIPAddressesDataSource,
		kubernetes.NewKubernetesClusterDataSource,
		loadbalancer.NewDNSChallengeDomainDataSource,
		managedobjectstorage.NewPoliciesDataSource,
		managedobjectstorage.NewRegionsDataSource,
		storage.NewStorageDataSource,
	}
}

func (p *upcloudProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		kubernetes.NewKubernetesClusterEphemeral,
	}
}
