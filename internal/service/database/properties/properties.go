package properties

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

func getPropertiesData(dbType upcloud.ManagedDatabaseServiceType) (string, error) {
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
		return "", fmt.Errorf(`unknown database type "%s"`, dbType)
	}
}

func getPropertiesMap(dbType upcloud.ManagedDatabaseServiceType) (map[string]upcloud.ManagedDatabaseServiceProperty, error) {
	var properties map[string]upcloud.ManagedDatabaseServiceProperty

	data, err := getPropertiesData(dbType)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(data), &properties)
	if err != nil {
		return nil, err
	}

	return properties, nil
}
