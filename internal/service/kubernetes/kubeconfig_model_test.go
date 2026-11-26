package kubernetes

import (
	"encoding/base64"
	"fmt"
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

const testKubeconfigTemplate = `apiVersion: v1
kind: Config
clusters:
- name: example
  cluster:
    certificate-authority-data: %s
    server: https://example.upcloudk8s.com:443
contexts:
- name: %s@example
  context:
    cluster: example
    user: %s
current-context: %s@example
users:
- name: %s
  user:
%s
`

func TestKubeconfigModel_parse(t *testing.T) {
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

	t.Run("cluster level", func(t *testing.T) {
		s := fmt.Sprintf(testKubeconfigTemplate, b64("CA"), "admin", "admin", "admin", "admin",
			"    client-certificate-data: "+b64("CERT")+"\n    client-key-data: "+b64("KEY"))

		m := kubeconfigModel{}
		name, diags := m.parse(s)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "example", name.ValueString())
		assert.Equal(t, s, m.Kubeconfig.ValueString())
		assert.Equal(t, "https://example.upcloudk8s.com:443", m.Host.ValueString())
		assert.Equal(t, "CA", m.ClusterCACertificate.ValueString())
		assert.Equal(t, "CERT", m.ClientCertificate.ValueString())
		assert.Equal(t, "KEY", m.ClientKey.ValueString())
	})

	t.Run("OIDC", func(t *testing.T) {
		s := fmt.Sprintf(testKubeconfigTemplate, b64("CA"), "oidc", "oidc", "oidc", "oidc", "    exec:\n      command: kubectl")

		m := kubeconfigModel{}
		_, diags := m.parse(s)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "https://example.upcloudk8s.com:443", m.Host.ValueString())
		assert.Equal(t, "CA", m.ClusterCACertificate.ValueString())
		assert.True(t, m.ClientCertificate.IsNull())
		assert.True(t, m.ClientKey.IsNull())
	})

	t.Run("invalid", func(t *testing.T) {
		valid := fmt.Sprintf(testKubeconfigTemplate, b64("CA"), "oidc", "oidc", "oidc", "oidc", "    exec:\n      command: kubectl")
		for name, s := range map[string]string{
			"empty":              "",
			"no current context": "apiVersion: v1\nkind: Config\n",
			"unknown context":    "current-context: no-at-sign\n",
			"unknown user":       fmt.Sprintf(testKubeconfigTemplate, b64("CA"), "oidc", "other", "oidc", "oidc", "    exec:\n      command: kubectl"),
			"invalid CA":         fmt.Sprintf(testKubeconfigTemplate, "***", "oidc", "oidc", "oidc", "oidc", "    exec:\n      command: kubectl"),
		} {
			m := kubeconfigModel{}
			_, diags := m.parse(s)
			assert.True(t, diags.HasError(), name)
		}

		m := kubeconfigModel{}
		_, diags := m.parse(valid)
		assert.False(t, diags.HasError())
	})
}
