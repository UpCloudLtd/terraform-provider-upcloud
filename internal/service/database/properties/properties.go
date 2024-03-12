package properties

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
)

//go:generate sh generate_types_data.sh

//go:embed service_types_data.json
var serviceTypesData []byte

var (
	onceParseServiceTypes sync.Once
	serviceTypes          map[string]upcloud.ManagedDatabaseType
)

func GetProperties(dbType upcloud.ManagedDatabaseServiceType) map[string]upcloud.ManagedDatabaseServiceProperty {
	onceParseServiceTypes.Do(func() {
		err := json.Unmarshal(serviceTypesData, &serviceTypes)
		if err != nil {
			panic(panicMessage("database", "map", err))
		}
	})

	return serviceTypes[string(dbType)].Properties
}
