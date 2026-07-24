package kubernetes

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestReconcileNodeGroupStorageEncryption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		current  types.String
		apiValue upcloud.StorageEncryption
		expected types.String
	}{
		{
			name:     "preserves configured none when api omits value",
			current:  types.StringValue(string(upcloud.StorageEncryptionNone)),
			apiValue: "",
			expected: types.StringValue(string(upcloud.StorageEncryptionNone)),
		},
		{
			name:     "uses api value when present",
			current:  types.StringValue(string(upcloud.StorageEncryptionNone)),
			apiValue: upcloud.StorageEncryptionDataAtRest,
			expected: types.StringValue(string(upcloud.StorageEncryptionDataAtRest)),
		},
		{
			name:     "ends up with null for omitted unconfigured create value",
			current:  types.StringUnknown(),
			apiValue: "",
			expected: types.StringNull(),
		},
		{
			name:     "preserves configured none when api omits value during import",
			current:  types.StringValue(string(upcloud.StorageEncryptionNone)),
			apiValue: "",
			expected: types.StringValue(string(upcloud.StorageEncryptionNone)),
		},
		{
			name:     "ends up with null for omitted unconfigured import value",
			current:  types.StringNull(),
			apiValue: "",
			expected: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := setNodeGroupStorageEncryption(tt.current, tt.apiValue)
			require.Equal(t, tt.expected, got)
		})
	}
}
