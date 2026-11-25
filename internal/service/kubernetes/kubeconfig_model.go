package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.yaml.in/yaml/v3"
)

const (
	kubeconfigIDDescription             = "ID of the kubeconfig. This is a combination of the cluster UUID, the name of the OIDC authentication, and the audience, separated with a `/`. The authentication and the audience are included only if they are defined."
	kubeconfigAuthenticationDescription = "Name of the OIDC authentication issuer configuration (see `upcloud_kubernetes_cluster_authentication` resource) to get the kubeconfig for. If left out, the cluster level kubeconfig is returned."
	kubeconfigAudienceDescription       = "Audience of the issuer to use as the OIDC client ID in the kubeconfig. Needs to match one of the audiences of the authentication issuer configuration. If left out, the first audience of the issuer is used. Can be defined only when `authentication` is defined."
	kubeconfigSummary                   = "Kubeconfig of a [Managed Kubernetes](https://upcloud.com/products/managed-kubernetes) cluster. Returns the cluster level kubeconfig, or if `authentication` is defined, a kubeconfig for users authenticating through the OIDC authentication issuer configuration of that name. The OIDC kubeconfigs use the `kubelogin` (`kubectl oidc-login`) plugin, which needs to be installed on the client, and do not include any long-lived credentials. "
)

type kubernetesKubeconfigModel struct {
	ID                   types.String `tfsdk:"id"`
	Cluster              types.String `tfsdk:"cluster"`
	Authentication       types.String `tfsdk:"authentication"`
	Audience             types.String `tfsdk:"audience"`
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
	Host                 types.String `tfsdk:"host"`
	Kubeconfig           types.String `tfsdk:"kubeconfig"`
}

func (m *kubernetesKubeconfigModel) marshalID() string {
	components := []string{m.Cluster.ValueString()}
	if authentication := m.Authentication.ValueString(); authentication != "" {
		components = append(components, authentication)
	}
	if audience := m.Audience.ValueString(); audience != "" {
		components = append(components, audience)
	}
	return utils.MarshalID(components...)
}

func nameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, resourceNameMaxLength),
		stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name should only contain lowercase alphanumeric characters and dashes (a-z, 0-9, -). Name should not start or end with a dash. Regular expresion used to check validation: %s", resourceNameRegexp)),
	}
}

// parseOIDC fills the model from the OIDC kubeconfig contents. Unlike the cluster level kubeconfig, the OIDC kubeconfig does not include client certificate.
func (m *kubernetesKubeconfigModel) parseOIDC(s string) (diags diag.Diagnostics) {
	if s == "" {
		diags.AddError(
			"Cluster kubeconfig is empty",
			"API returned an empty kubeconfig for the cluster.",
		)
		return
	}

	m.Kubeconfig = types.StringValue(s)
	m.ClientCertificate = types.StringNull()
	m.ClientKey = types.StringNull()
	m.Host = types.StringNull()
	m.ClusterCACertificate = types.StringNull()

	k := kubeconfig{}
	if err := yaml.Unmarshal([]byte(s), &k); err != nil {
		diags.AddError(
			"Kubeconfig YAML unmarshal failed",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	clusterName := ""
	for _, c := range k.Contexts {
		if c.Name == k.CurrentContext {
			clusterName = c.Context.Cluster
		}
	}

	for _, c := range k.Clusters {
		if c.Name != clusterName {
			continue
		}

		m.Host = types.StringValue(c.Cluster.Server)
		if c.Cluster.CertificateAuthorityData != "" {
			decoded, err := base64.StdEncoding.DecodeString(c.Cluster.CertificateAuthorityData)
			if err != nil {
				diags.AddError(
					"Unable to decode base64 encoded value of cluster CA certificate",
					utils.ErrorDiagnosticDetail(err),
				)
				return
			}
			m.ClusterCACertificate = types.StringValue(string(decoded))
		}
	}

	return
}

// parseCluster fills the model from the cluster level kubeconfig contents.
func (m *kubernetesKubeconfigModel) parseCluster(s string) diag.Diagnostics {
	var parsed kubeconfigModel
	_, diags := parsed.parse(s)
	if diags.HasError() {
		return diags
	}

	m.ClientCertificate = parsed.ClientCertificate
	m.ClientKey = parsed.ClientKey
	m.ClusterCACertificate = parsed.ClusterCACertificate
	m.Host = parsed.Host
	m.Kubeconfig = parsed.Kubeconfig
	return diags
}

// fetch gets the kubeconfig from the API and sets the kubeconfig data and ID to the model.
func (m *kubernetesKubeconfigModel) fetch(ctx context.Context, client *v9.ClientWithResponses) (diags diag.Diagnostics) {
	clusterUUID, d := parseClusterUUID(m.Cluster.ValueString())
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	var s string
	if m.Authentication.IsNull() {
		s, d = fetchClusterKubeconfig(ctx, client, clusterUUID)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		diags.Append(m.parseCluster(s)...)
	} else {
		s, d = m.fetchOIDCKubeconfig(ctx, client, clusterUUID)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		diags.Append(m.parseOIDC(s)...)
	}

	if !diags.HasError() {
		m.ID = types.StringValue(m.marshalID())
	}
	return diags
}

func fetchClusterKubeconfig(ctx context.Context, client *v9.ClientWithResponses, clusterUUID uuid.UUID) (string, diag.Diagnostics) {
	apiResp, err := client.GetKubernetesClusterKubeconfigWithResponse(ctx, clusterUUID)
	if err != nil {
		return "", clusterAuthenticationError("Unable to read cluster kubeconfig", err)
	}
	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		return "", clusterAuthenticationStatusError("Unable to read cluster kubeconfig", apiResp.Status(), apiResp.Body)
	}
	return apiResp.JSON200.Kubeconfig, nil
}

func (m *kubernetesKubeconfigModel) fetchOIDCKubeconfig(ctx context.Context, client *v9.ClientWithResponses, clusterUUID uuid.UUID) (string, diag.Diagnostics) {
	var params *v9.GetKubernetesClusterOIDCKubeconfigParams
	if !m.Audience.IsNull() {
		audience := m.Audience.ValueString()
		params = &v9.GetKubernetesClusterOIDCKubeconfigParams{Audience: &audience}
	}

	apiResp, err := client.GetKubernetesClusterOIDCKubeconfigWithResponse(ctx, clusterUUID, m.Authentication.ValueString(), params)
	if err != nil {
		return "", clusterAuthenticationError("Unable to read cluster kubeconfig", err)
	}
	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		return "", clusterAuthenticationStatusError("Unable to read cluster kubeconfig", apiResp.Status(), apiResp.Body)
	}
	return apiResp.JSON200.Kubeconfig, nil
}
