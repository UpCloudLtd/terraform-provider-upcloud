package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type kubeconfigModel struct {
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
	Host                 types.String `tfsdk:"host"`
	Kubeconfig           types.String `tfsdk:"kubeconfig"`
}

var kubeconfigAttrTypes = map[string]attr.Type{
	"client_certificate":     types.StringType,
	"client_key":             types.StringType,
	"cluster_ca_certificate": types.StringType,
	"host":                   types.StringType,
	"kubeconfig":             types.StringType,
}

// setB64Decoded sets the field to the decoded value. Empty values are left unset, as for example OIDC kubeconfigs do not include client certificate.
func (m *kubeconfigModel) setB64Decoded(field string, encoded string) (diags diag.Diagnostics) {
	if encoded == "" {
		return diags
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Unable to decode base64 encoded value of %s", field),
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	value := types.StringValue(string(decoded))
	reflect.ValueOf(m).Elem().FieldByName(field).Set(reflect.ValueOf(value))
	return diags
}

// parse fills the model from the cluster level or OIDC kubeconfig contents and returns the cluster name found from the kubeconfig.
func (m *kubeconfigModel) parse(s string) (name types.String, diags diag.Diagnostics) {
	name = types.StringNull()
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
	m.ClusterCACertificate = types.StringNull()
	m.Host = types.StringNull()

	k := kubeconfig{}
	err := yaml.Unmarshal([]byte(s), &k)
	if err != nil {
		diags.AddError(
			"Kubeconfig YAML unmarshal failed",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	// Resolve the cluster and user of the current context.
	var contextData *kubeconfigContextData
	for _, v := range k.Contexts {
		if v.Name == k.CurrentContext {
			contextData = &v.Context
			break
		}
	}
	if contextData == nil {
		diags.AddError(
			"Unexpected kubeconfig current context",
			fmt.Sprintf("Unable to find the current context %q from the kubeconfig contexts.", k.CurrentContext),
		)
		return
	}

	foundCluster := false
	for _, v := range k.Clusters {
		if v.Name == contextData.Cluster {
			foundCluster = true
			diags.Append(m.setB64Decoded("ClusterCACertificate", v.Cluster.CertificateAuthorityData)...)
			m.Host = types.StringValue(v.Cluster.Server)
		}
	}
	if !foundCluster {
		diags.AddError(
			"Unexpected kubeconfig contents",
			fmt.Sprintf("Unable to find the cluster %q of the current context from the kubeconfig clusters.", contextData.Cluster),
		)
		return
	}

	foundUser := false
	for _, v := range k.Users {
		if v.Name == contextData.User {
			foundUser = true
			diags.Append(m.setB64Decoded("ClientCertificate", v.User.ClientCertificateData)...)
			diags.Append(m.setB64Decoded("ClientKey", v.User.ClientKeyData)...)
		}
	}
	if !foundUser {
		diags.AddError(
			"Unexpected kubeconfig contents",
			fmt.Sprintf("Unable to find the user %q of the current context from the kubeconfig users.", contextData.User),
		)
		return
	}

	return types.StringValue(contextData.Cluster), diags
}

type kubernetesKubeconfigModel struct {
	ID             types.String `tfsdk:"id"`
	Cluster        types.String `tfsdk:"cluster"`
	Authentication types.String `tfsdk:"authentication"`
	Audience       types.String `tfsdk:"audience"`

	kubeconfigModel
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
	} else {
		s, d = m.fetchOIDCKubeconfig(ctx, client, clusterUUID)
	}
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	_, d = m.parse(s)
	diags.Append(d...)

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
