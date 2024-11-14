package properties

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/stretchr/testify/assert"
)

func getTypes() []upcloud.ManagedDatabaseServiceType {
	return []upcloud.ManagedDatabaseServiceType{
		upcloud.ManagedDatabaseServiceTypeMySQL,
		upcloud.ManagedDatabaseServiceTypeOpenSearch,
		upcloud.ManagedDatabaseServiceTypePostgreSQL,
		upcloud.ManagedDatabaseServiceTypeRedis, //nolint:staticcheck // To be removed when Redis support has been removed
		upcloud.ManagedDatabaseServiceTypeValkey,
	}
}

func TestGetPropertiesMap(t *testing.T) {
	dbTypes := getTypes()

	for _, dbType := range dbTypes {
		t.Run(string(dbType), func(t *testing.T) {
			assert.NotPanics(t, func() {
				GetProperties(dbType)
			})
		})
	}
}
