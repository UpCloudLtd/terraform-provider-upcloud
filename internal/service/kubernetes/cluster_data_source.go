package kubernetes

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewKubernetesClusterDataSource() datasource.DataSource {
	return &kubernetesClusterDataSource{}
}

var (
	_ datasource.DataSource              = &kubernetesClusterDataSource{}
	_ datasource.DataSourceWithConfigure = &kubernetesClusterDataSource{}
)

type kubernetesClusterDataSource struct {
	client *service.Service
}

func (d *kubernetesClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (d *kubernetesClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type kubernetesClusterDataModel struct {
	kubeconfigModel

	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m *kubernetesClusterDataModel) setKubeconfigData(s string) (diags diag.Diagnostics) {
	name, diags := m.parse(s)
	if !name.IsNull() {
		m.Name = name
	}
	return diags
}

func (d *kubernetesClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *kubernetesClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kubernetesClusterDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	clusterID := data.ID.ValueString()

	s, err := d.client.GetKubernetesKubeconfig(ctx, &request.GetKubernetesKubeconfigRequest{
		UUID: clusterID,
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read cluster kubeconfig",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(data.setKubeconfigData(s)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
