package kubernetes

import (
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubernetesKubeconfigModel_marshalID(t *testing.T) {
	m := kubernetesKubeconfigModel{
		Cluster: types.StringValue("0a1b2c3d-0000-0000-0000-000000000000"),
	}
	assert.Equal(t, "0a1b2c3d-0000-0000-0000-000000000000", m.marshalID())

	m.Authentication = types.StringValue("issuer")
	assert.Equal(t, "0a1b2c3d-0000-0000-0000-000000000000/issuer", m.marshalID())

	m.Audience = types.StringValue("client-id")
	assert.Equal(t, "0a1b2c3d-0000-0000-0000-000000000000/issuer/client-id", m.marshalID())
}

func TestKubernetesKubeconfigModel_parseOIDC(t *testing.T) {
	ca := base64.StdEncoding.EncodeToString([]byte("CA"))
	s := `apiVersion: v1
kind: Config
clusters:
- name: example
  cluster:
    certificate-authority-data: ` + ca + `
    server: https://example.upcloudk8s.com:443
contexts:
- name: oidc@example
  context:
    cluster: example
    user: oidc
current-context: oidc@example
users:
- name: oidc
  user:
    exec:
      command: kubectl
`

	m := kubernetesKubeconfigModel{}
	diags := m.parseOIDC(s)
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, s, m.Kubeconfig.ValueString())
	assert.Equal(t, "https://example.upcloudk8s.com:443", m.Host.ValueString())
	assert.Equal(t, "CA", m.ClusterCACertificate.ValueString())

	assert.True(t, (&kubernetesKubeconfigModel{}).parseOIDC("").HasError())
}

func TestKubeconfigModel_parseInvalidCurrentContext(t *testing.T) {
	for _, s := range []string{"apiVersion: v1\nkind: Config\n", "current-context: no-at-sign\n"} {
		m := kubernetesKubeconfigModel{}
		assert.True(t, m.parseCluster(s).HasError(), s)
	}
}
