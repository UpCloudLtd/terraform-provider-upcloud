package loadbalancer

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDNSChallengeDomainDataSource() datasource.DataSource {
	return &dnsChallengeDomainDataSource{}
}

var (
	_ datasource.DataSource              = &dnsChallengeDomainDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsChallengeDomainDataSource{}
)

type dnsChallengeDomainDataSource struct {
	client *service.Service
}

func (d *dnsChallengeDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_dns_challenge_domain"
}

func (d *dnsChallengeDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type regionsModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
}

func (d *dnsChallengeDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns DNS challenge domain to use when validating domain ownership using ACME challenge method.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"domain": schema.StringAttribute{
				Description: "The domain to use for DNS challenge validation.",
				Computed:    true,
			},
		},
	}
}

func (d *dnsChallengeDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data regionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	data.ID = types.StringValue(time.Now().UTC().String())

	domain, err := d.client.GetLoadBalancerDNSChallengeDomain(ctx, &request.GetLoadBalancerDNSChallengeDomainRequest{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read DNS challenge domain",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.Domain = types.StringValue(domain.Domain)
	data.ID = types.StringValue(domain.Domain)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
