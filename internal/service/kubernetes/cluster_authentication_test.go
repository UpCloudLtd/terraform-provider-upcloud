package kubernetes

import (
	"context"
	"testing"

	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClusterAuthenticationID(t *testing.T) {
	uuid := "0a1b2c3d-0000-4000-8000-000000000000"

	clusterUUID, name, diags := parseClusterAuthenticationID(uuid + "/issuer")
	assert.False(t, diags.HasError())
	assert.Equal(t, uuid, clusterUUID.String())
	assert.Equal(t, "issuer", name)

	for _, id := range []string{"", uuid, uuid + "/", "not-a-uuid/issuer", uuid + "/issuer/extra"} {
		_, _, diags := parseClusterAuthenticationID(id)
		assert.True(t, diags.HasError(), id)
	}
}

func TestClusterAuthentication_setValuesBuildRequest(t *testing.T) {
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	(&kubernetesClusterAuthenticationResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)

	claim, prefix, emptyPrefix := "email", "oidc:", ""
	hd, hdValue := "hd", "example.com"
	expression, message := "claims.email_verified == true", "Email must be verified."
	api := v9.KubernetesAuthenticationIssuerConfig{
		Name:      "issuer",
		IssuerUrl: "https://accounts.example.com",
		Audiences: []string{"a", "b"},
		ClaimMappings: &v9.KubernetesClaimMappings{
			Username: &v9.KubernetesPrefixedClaimOrExpression{Claim: &claim, Prefix: &prefix},
			Groups:   &v9.KubernetesPrefixedClaimOrExpression{Claim: &claim, Prefix: &emptyPrefix},
			Extra:    &[]v9.KubernetesExtraMapping{{Key: "example.org/foo", ValueExpression: "claims.foo"}},
		},
		ClaimValidationRules: &[]v9.KubernetesClaimValidationRule{
			{Claim: &hd, RequiredValue: &hdValue},
			{Expression: &expression, Message: &message},
		},
		UserValidationRules: &[]v9.KubernetesUserValidationRule{{Expression: "true"}},
	}

	m := clusterAuthenticationModel{Cluster: types.StringValue("0a1b2c3d-0000-4000-8000-000000000000")}
	require.False(t, setClusterAuthenticationValues(ctx, &m, &api).HasError())

	// Setting and getting the state verifies that the models match the schema.
	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, &m).HasError())
	var got clusterAuthenticationModel
	require.False(t, state.Get(ctx, &got).HasError())

	config, diags := buildClusterAuthenticationRequest(ctx, got)
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, api.Audiences, config.Audiences)
	assert.Equal(t, api.ClaimMappings, config.ClaimMappings)
	assert.Equal(t, api.ClaimValidationRules, config.ClaimValidationRules)
	assert.Equal(t, api.UserValidationRules, config.UserValidationRules)

	// Optional attributes are null when the API does not return them.
	minimal := clusterAuthenticationModel{Cluster: m.Cluster}
	require.False(t, setClusterAuthenticationValues(ctx, &minimal, &v9.KubernetesAuthenticationIssuerConfig{Name: "issuer", Audiences: []string{"a"}}).HasError())
	assert.True(t, minimal.ClaimMappings.IsNull())
	assert.True(t, minimal.ClaimValidationRules.IsNull())
	assert.True(t, minimal.UserValidationRules.IsNull())
}
