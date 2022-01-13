package upcloudschema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
)

var (
	//go:embed managed_database_service_types.json
	rawManagedDatabaseSchema []byte
)

var (
	onceParseManagedDatabaseSchema sync.Once
	managedDatabaseServiceTypes    map[upcloud.ManagedDatabaseServiceType]map[string]interface{}
)

// ManagedDatabaseServiceTypeSchema parses the schema once and then caches it. The requested service type schema
// is then returned. All errors panic
func ManagedDatabaseServiceTypeSchema(serviceType upcloud.ManagedDatabaseServiceType) map[string]interface{} {
	onceParseManagedDatabaseSchema.Do(func() {
		if err := json.Unmarshal(rawManagedDatabaseSchema, &managedDatabaseServiceTypes); err != nil {
			panic(fmt.Sprintf("managed database service types schema load failure: %v", err))
		}
	})
	return managedDatabaseServiceTypes[serviceType]
}

// ManagedDatabaseServicePropertiesSchema returns the service's properties schema. It calls ManagedDatabaseServiceTypeSchema
// so every error will be a panic
func ManagedDatabaseServicePropertiesSchema(serviceType upcloud.ManagedDatabaseServiceType) map[string]interface{} {
	svcSchema := ManagedDatabaseServiceTypeSchema(serviceType)
	if _, ok := svcSchema["properties"]; !ok {
		return nil
	}
	return svcSchema["properties"].(map[string]interface{})
}
