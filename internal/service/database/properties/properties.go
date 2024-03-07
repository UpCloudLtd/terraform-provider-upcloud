package properties

import (
	_ "embed"
	"encoding/json"
	"fmt"

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

func getPropertiesData(dbType upcloud.ManagedDatabaseServiceType) ([]byte, error) {
	switch dbType {
	case upcloud.ManagedDatabaseServiceTypeMySQL:
		return mysqlPropertiesJSON, nil
	case upcloud.ManagedDatabaseServiceTypeOpenSearch:
		return opensearchPropertiesJSON, nil
	case upcloud.ManagedDatabaseServiceTypePostgreSQL:
		return pgPropertiesJSON, nil
	case upcloud.ManagedDatabaseServiceTypeRedis:
		return redisPropertiesJSON, nil
	default:
		return nil, fmt.Errorf(`unknown database type "%s"`, dbType)
	}
}

func getPropertiesMap(dbType upcloud.ManagedDatabaseServiceType) (map[string]upcloud.ManagedDatabaseServiceProperty, error) {
	var properties map[string]upcloud.ManagedDatabaseServiceProperty

	data, err := getPropertiesData(dbType)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &properties)
	if err != nil {
		return nil, err
	}

	return properties, nil
}
