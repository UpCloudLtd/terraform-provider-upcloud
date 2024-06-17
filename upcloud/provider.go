package upcloud

import (
	"context"
	"os"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/cloud"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/ip"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/managedobjectstorage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/network"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/networkpeering"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/router"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	usernameDescription       = "UpCloud username with API access. Can also be configured using the `UPCLOUD_USERNAME` environment variable."
	passwordDescription       = "Password for UpCloud API user. Can also be configured using the `UPCLOUD_PASSWORD` environment variable."
	requestTimeoutDescription = "The duration (in seconds) that the provider waits for an HTTP request towards UpCloud API to complete. Defaults to 120 seconds"
)

type upcloudProviderModel struct {
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	RetryWaitMinSec   types.Int64  `tfsdk:"retry_wait_min_sec"`
	RetryWaitMaxSec   types.Int64  `tfsdk:"retry_wait_max_sec"`
	RetryMax          types.Int64  `tfsdk:"retry_max"`
	RequestTimeoutSec types.Int64  `tfsdk:"request_timeout_sec"`
}

type upcloudProvider struct{}

var _ provider.Provider = New()

func New() provider.Provider {
	return &upcloudProvider{}
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

func withDefault[T comparable](val T, def T) T {
	var empty T
	if val == empty {
		return def
	}
	return val
}

func withEnvDefault(val string, env string) string {
	return withDefault(val, os.Getenv(env))
}

func (p *upcloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model upcloudProviderModel
	if diags := req.Config.Get(ctx, &model); diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	requestTimeout := time.Duration(withDefault(model.RequestTimeoutSec.ValueInt64(), 120)) * time.Second
	config := Config{
		Username: withEnvDefault(model.Username.ValueString(), "UPCLOUD_USERNAME"),
		Password: withEnvDefault(model.Password.ValueString(), "UPCLOUD_PASSWORD"),
	}

	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = time.Duration(withDefault(model.RetryWaitMinSec.ValueInt64(), 1)) * time.Second
	httpClient.RetryWaitMax = time.Duration(withDefault(model.RetryWaitMaxSec.ValueInt64(), 30)) * time.Second
	httpClient.RetryMax = int(withDefault(model.RetryMax.ValueInt64(), 4))

	service := newUpCloudServiceConnection(
		config.Username,
		config.Password,
		httpClient.HTTPClient,
		requestTimeout,
	)

	_, err := config.checkLogin(service)
	if err != nil {
		resp.Diagnostics.AddError("Authentication failed", "Failed to authenticate to UpCloud API with given credentials")
	}

	resp.ResourceData = service
	resp.DataSourceData = service
}

func (p *upcloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ip.NewFloatingIPAddressResource,
		network.NewNetworkResource,
		networkpeering.NewNetworkPeeringResource,
		router.NewRouterResource,
	}
}

func (p *upcloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloud.NewZoneDataSource,
		managedobjectstorage.NewRegionsDataSource,
	}
}
