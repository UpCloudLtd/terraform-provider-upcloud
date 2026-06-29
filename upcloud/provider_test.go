package upcloud

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func toTfTypesValue(t *testing.T, v attr.Value) tftypes.Value {
	val, err := v.ToTerraformValue(context.Background())
	require.NoError(t, err)
	return val
}

func TestProvider_LoggingAndUserAgent(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping because TF_ACC is not set. This test sends requests to UpCloud API and requires valid credentials.")
	}

	t.Parallel()

	tests := []struct {
		name              string
		assertContains    []string
		assertNotContains []string
		provider          provider.Provider
	}{
		{
			name:              "default user agent",
			assertContains:    []string{config.DefaultUserAgent()},
			assertNotContains: []string{"custom-user-agent"},
			provider:          New(),
		},
		{
			name:              "custom user agent",
			assertContains:    []string{"custom-user-agent"},
			assertNotContains: []string{config.DefaultUserAgent()},
			provider:          NewWithUserAgent("custom-user-agent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw := tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"username":            tftypes.String,
					"password":            tftypes.String,
					"token":               tftypes.String,
					"retry_wait_min_sec":  tftypes.Number,
					"retry_wait_max_sec":  tftypes.Number,
					"retry_max":           tftypes.Number,
					"request_timeout_sec": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"username":            toTfTypesValue(t, types.StringNull()),
				"password":            toTfTypesValue(t, types.StringNull()),
				"token":               toTfTypesValue(t, types.StringNull()),
				"retry_wait_min_sec":  toTfTypesValue(t, types.NumberNull()),
				"retry_wait_max_sec":  toTfTypesValue(t, types.NumberNull()),
				"retry_max":           toTfTypesValue(t, types.NumberNull()),
				"request_timeout_sec": toTfTypesValue(t, types.NumberNull()),
			})

			var out strings.Builder
			ctx := tflogtest.RootLogger(context.Background(), &out)

			schemaResp := provider.SchemaResponse{}
			tt.provider.Schema(ctx, provider.SchemaRequest{}, &schemaResp)

			req := provider.ConfigureRequest{
				Config: tfsdk.Config{
					Raw:    raw,
					Schema: schemaResp.Schema,
				},
			}

			resp := provider.ConfigureResponse{}
			tt.provider.Configure(ctx, req, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("provider configuration failed: %v", resp.Diagnostics)
			}

			client, diags := utils.GetClientFromProviderData(resp.ResourceData)
			if diags.HasError() {
				t.Fatalf("failed to get client from provider data: %v", diags)
			}

			v9Client, diags := utils.GetV9ClientFromProviderData(resp.ResourceData)
			if diags.HasError() {
				t.Fatalf("failed to get v9 client from provider data: %v", diags)
			}

			_, err := client.GetManagedObjectStorages(ctx, &request.GetManagedObjectStoragesRequest{})
			assert.NoError(t, err)

			_, err = v9Client.ListObjectStoragesWithResponse(ctx, &upcloud.ListObjectStoragesParams{})
			assert.NoError(t, err)

			output := out.String()
			for _, expected := range tt.assertContains {
				assert.Contains(t, output, expected)
			}
			for _, notExpected := range tt.assertNotContains {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}
