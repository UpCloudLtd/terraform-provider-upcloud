package upcloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
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
			"upcloud_object_storage":      resourceUpCloudObjectStorage(),
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
	httpClient.RetryWaitMin = time.Duration(d.Get("retry_wait_min_sec").(int) * int(time.Second))
	httpClient.RetryWaitMax = time.Duration(d.Get("retry_wait_max_sec").(int) * int(time.Second))
	httpClient.RetryMax = d.Get("retry_max").(int)

	service := newUpCloudServiceConnection(
		d.Get("username").(string),
		d.Get("password").(string),
		logging(httpClient.HTTPClient),
	)

	_, err := config.checkLogin(service)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return service, diags
}

type loggingRoundTripper struct {
	client *http.Client
}

type closeBuffer struct {
	*bytes.Buffer
}

func (c closeBuffer) Close() error {
	return nil
}

func (l loggingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var postBody []byte
	if request.Method == "POST" {
		postBody, request.Body = readAndCloneReader(request.Body)
	}
	res, err := l.client.Transport.RoundTrip(request)
	_, _ = fmt.Fprintf(os.Stdout, "RTRIP\n%v %v\n", request.Method, request.URL)
	if len(postBody) > 0 {
		_, _ = fmt.Fprintln(os.Stdout, string(postBody))
	}
	var responseBody []byte
	newRes := *res
	responseBody, newRes.Body = readAndCloneReader(res.Body)
	if res.StatusCode >= 300 {
		_, _ = fmt.Fprintf(os.Stdout, "%v %v\n%v\n%v\n", res.StatusCode, res.Status, string(responseBody), err)
	}
	return &newRes, err
}

func readAndCloneReader(body io.ReadCloser) ([]byte, io.ReadCloser) {
	read, _ := ioutil.ReadAll(body)
	return read, closeBuffer{bytes.NewBuffer(read)}
}

func logging(httpClient *http.Client) *http.Client {
	loggingTransport := loggingRoundTripper{client: httpClient}
	return &http.Client{
		Transport: loggingTransport,
	}
}

func newUpCloudServiceConnection(username, password string, httpClient *http.Client) *service.Service {
	client := client.NewWithHTTPClient(username, password, httpClient)
	client.UserAgent = fmt.Sprintf("terraform-provider-upcloud/%s", config.Version)
	client.SetTimeout(upcloudAPITimeout)

	return service.New(client)
}
