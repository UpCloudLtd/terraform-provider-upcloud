package kubernetes

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NewKubernetesKubeconfigEphemeral() ephemeral.EphemeralResource {
	return &kubernetesKubeconfigEphemeral{}
}

var (
	_ ephemeral.EphemeralResource              = &kubernetesKubeconfigEphemeral{}
	_ ephemeral.EphemeralResourceWithConfigure = &kubernetesKubeconfigEphemeral{}
)

type kubernetesKubeconfigEphemeral struct {
	client *v9.ClientWithResponses
}

func (e *kubernetesKubeconfigEphemeral) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_kubeconfig"
}

func (e *kubernetesKubeconfigEphemeral) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	e.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

func (e *kubernetesKubeconfigEphemeral) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
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

func (e *kubernetesKubeconfigEphemeral) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data kubernetesKubeconfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.fetch(ctx, e.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
