package properties

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetSchema(t *testing.T) {
	tests := []struct {
		db       upcloud.ManagedDatabaseServiceType
		key      string
		computed bool
		defaultf func(interface{}) bool
	}{
		{
			db:       "pg",
			key:      "public_access",
			computed: false,
			defaultf: func(v interface{}) bool { return v.(bool) == false },
		},
		{
			db:       "redis",
			key:      "redis_notify_keyspace_events",
			computed: false,
		},
	}

	dbTypes := getTypes()
	for _, dbType := range dbTypes {
		var schema map[string]*schema.Schema
		t.Run(string(dbType), func(t *testing.T) {
			assert.NotPanics(t, func() {
				schema = GetSchemaMap(dbType)
			})

			for _, test := range tests {
				if test.db == dbType {
					t.Run(fmt.Sprintf("%s", test.key), func(t *testing.T) {
						s := schema[test.key]

						assert.Equal(t, test.computed, s.Computed)
						if test.defaultf != nil {
							assert.True(t, test.defaultf(s.Default))
						}
					})
				}
			}
		})
	}
}
