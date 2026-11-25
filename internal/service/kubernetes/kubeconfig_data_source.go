package kubernetes

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NewKubernetesKubeconfigDataSource() datasource.DataSource {
	return &kubernetesKubeconfigDataSource{}
}

var (
	_ datasource.DataSource              = &kubernetesKubeconfigDataSource{}
	_ datasource.DataSourceWithConfigure = &kubernetesKubeconfigDataSource{}
)

type kubernetesKubeconfigDataSource struct {
	client *v9.ClientWithResponses
}

func (d *kubernetesKubeconfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_kubeconfig"
}

func (d *kubernetesKubeconfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (d *kubernetesKubeconfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: kubeconfigSummary + "Please refer to [Terraform documentation on sensitive data](https://www.terraform.io/language/state/sensitive-data) to keep the kubeconfig as safe as possible.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: kubeconfigIDDescription,
			},
			"cluster": schema.StringAttribute{
				Required:    true,
				Description: idDescription,
			},
			"authentication": schema.StringAttribute{
				Optional:    true,
				Description: kubeconfigAuthenticationDescription,
				Validators:  nameValidators(),
			},
			"audience": schema.StringAttribute{
				Optional:    true,
				Description: kubeconfigAudienceDescription,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("authentication")),
				},
			},
			"client_certificate": schema.StringAttribute{
				Description: clientCertificateDescription + " Not set for OIDC kubeconfigs.",
				Computed:    true,
			},
			"client_key": schema.StringAttribute{
				Description: clientKeyDescription + " Not set for OIDC kubeconfigs.",
				Computed:    true,
				Sensitive:   true,
			},
			"cluster_ca_certificate": schema.StringAttribute{
				Description: clusterCACertificateDescription,
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: hostDescription,
			},
			"kubeconfig": schema.StringAttribute{
				Description: kubeconfigDescription,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (d *kubernetesKubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kubernetesKubeconfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.fetch(ctx, d.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
