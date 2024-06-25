package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
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
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
	ID                   types.String `tfsdk:"id"`
	Host                 types.String `tfsdk:"host"`
	Kubeconfig           types.String `tfsdk:"kubeconfig"`
	Name                 types.String `tfsdk:"name"`
}

func (m *kubernetesClusterDataModel) setB64Decoded(field string, encoded string) (diags diag.Diagnostics) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Unable to decode base64 encoded value of %s", field),
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	value := types.StringValue(string(decoded))
	reflect.ValueOf(m).Elem().FieldByName(field).Set(reflect.ValueOf(value))
	return
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
				Computed:    true,
				Description: kubeconfigDescription,
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
	if s == "" {
		resp.Diagnostics.AddError(
			"Cluster kubeconfig is empty",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.Kubeconfig = types.StringValue(s)

	k := kubeconfig{}
	err = yaml.Unmarshal([]byte(s), &k)
	if err != nil {
		resp.Diagnostics.AddError(
			"Kubeconfig YAML unmarshal failed",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	currentContext := strings.Split(k.CurrentContext, "@")
	data.Name = types.StringValue(currentContext[1])

	for _, v := range k.Clusters {
		if v.Name == currentContext[1] {
			resp.Diagnostics.Append(data.setB64Decoded("ClusterCACertificate", v.Cluster.CertificateAuthorityData)...)
			data.Host = types.StringValue(v.Cluster.Server)
		}
	}

	for _, v := range k.Users {
		if v.Name == currentContext[0] {
			resp.Diagnostics.Append(data.setB64Decoded("ClientCertificate", v.User.ClientCertificateData)...)
			resp.Diagnostics.Append(data.setB64Decoded("ClientKey", v.User.ClientKeyData)...)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
