package properties

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

//go:generate sh generate_types_data.sh

//go:embed mysql_properties.json
var mysqlPropertiesJSON []byte

//go:embed opensearch_properties.json
var opensearchPropertiesJSON []byte

//go:embed pg_properties.json
var pgPropertiesJSON []byte

//go:embed redis_properties.json
var redisPropertiesJSON []byte

var (
	onceParseProperties sync.Once
	propertiesByType    map[upcloud.ManagedDatabaseServiceType]map[string]upcloud.ManagedDatabaseServiceProperty
)

func getTypes() []upcloud.ManagedDatabaseServiceType {
	return []upcloud.ManagedDatabaseServiceType{
		upcloud.ManagedDatabaseServiceTypeMySQL,
		upcloud.ManagedDatabaseServiceTypeOpenSearch,
		upcloud.ManagedDatabaseServiceTypePostgreSQL,
		upcloud.ManagedDatabaseServiceTypeRedis,
	}
}

func getPropertiesData(dbType upcloud.ManagedDatabaseServiceType) []byte {
	switch dbType {
	case upcloud.ManagedDatabaseServiceTypeMySQL:
		return mysqlPropertiesJSON
	case upcloud.ManagedDatabaseServiceTypeOpenSearch:
		return opensearchPropertiesJSON
	case upcloud.ManagedDatabaseServiceTypePostgreSQL:
		return pgPropertiesJSON
	case upcloud.ManagedDatabaseServiceTypeRedis:
		return redisPropertiesJSON
	default:
		return nil
	}
}

func parsePropertiesMap() {
	// This Function needs some extra logic because the JSON produced by upctl is not compatible with JSON provided by API because of the missing zone layer from ".pg.service_plans[].zones"
	propsMap := make(map[upcloud.ManagedDatabaseServiceType]map[string]upcloud.ManagedDatabaseServiceProperty)
	for _, dbType := range getTypes() {
		var props map[string]upcloud.ManagedDatabaseServiceProperty
		data := getPropertiesData(dbType)

		err := json.Unmarshal(data, &props)
		if err != nil {
			panic(panicMessage(dbType, "map", err))
		}

		propsMap[dbType] = props
	}

	propertiesByType = propsMap
}

func GetProperties(dbType upcloud.ManagedDatabaseServiceType) map[string]upcloud.ManagedDatabaseServiceProperty {
	onceParseProperties.Do(parsePropertiesMap)

	return propertiesByType[dbType]
}
