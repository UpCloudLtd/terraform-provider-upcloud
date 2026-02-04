package kubernetes

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
)

func NewKubernetesClusterEphemeral() ephemeral.EphemeralResource {
	return &kubernetesClusterEphemeral{}
}

var (
	_ ephemeral.EphemeralResource              = &kubernetesClusterEphemeral{}
	_ ephemeral.EphemeralResourceWithConfigure = &kubernetesClusterEphemeral{}
)

type kubernetesClusterEphemeral struct {
	client *service.Service
}

func (d *kubernetesClusterEphemeral) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (d *kubernetesClusterEphemeral) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

func (d *kubernetesClusterEphemeral) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Managed Kubernetes cluster details. Please refer to [Terraform documentation on sensitive data](https://www.terraform.io/language/state/sensitive-data) to keep the credential data as safe as possible.",
		Attributes: map[string]schema.Attribute{
			"client_certificate": schema.StringAttribute{
				Description: clientCertificateDescription,
				Computed:    true,
			},
			"client_key": schema.StringAttribute{
				Description: clientKeyDescription,
				Computed:    true,
				Sensitive:   true,
			},
			"cluster_ca_certificate": schema.StringAttribute{
				Description: clusterCACertificateDescription,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Required:    true,
				Description: idDescription,
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
			"name": schema.StringAttribute{
				Computed:    true,
				Description: nameDescription,
			},
		},
	}
}

func (d *kubernetesClusterEphemeral) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data kubernetesClusterDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	clusterID := data.ID.ValueString()

	s, err := d.client.GetKubernetesKubeconfig(ctx, &request.GetKubernetesKubeconfigRequest{
		UUID: clusterID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read cluster kubeconfig",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setClusterKubeconfigData(ctx, s, &data)...)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
