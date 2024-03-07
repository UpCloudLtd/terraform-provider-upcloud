package database

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestWaitServiceNameToPropagate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err != nil {
		t.Errorf("waitServiceNameToPropagate failed %+v", err)
	}
}

func TestWaitServiceNameToPropagateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	name := "upcloud.com"
	if err := waitServiceNameToPropagate(ctx, name); err == nil {
		d, _ := ctx.Deadline()
		t.Errorf("waitServiceNameToPropagate failed didn't timeout before deadline %s", d.Format(time.RFC3339))
	}
}

func TestDatabaseProperties(t *testing.T) {
	dbTypes := []upcloud.ManagedDatabaseServiceType{
		upcloud.ManagedDatabaseServiceTypeMySQL,
		upcloud.ManagedDatabaseServiceTypeOpenSearch,
		upcloud.ManagedDatabaseServiceTypePostgreSQL,
		upcloud.ManagedDatabaseServiceTypeRedis,
	}

	for _, dbType := range dbTypes {
		t.Run(string(dbType), func(t *testing.T) {
			s := properties.GetSchemaMap(dbType)
			testProperties(t, string(dbType), s)
		})
	}
}

func testProperties(t *testing.T, dbType string, s map[string]*schema.Schema) {
	username := os.Getenv("UPCLOUD_USERNAME")
	password := os.Getenv("UPCLOUD_PASSWORD")
	if username == "" || password == "" {
		t.Skip("UpCloud credentials not set.")
	}
	svc := service.New(client.New(username, password))
	dbt, err := svc.GetManagedDatabaseServiceType(context.Background(), &request.GetManagedDatabaseServiceTypeRequest{
		Type: dbType,
	})
	if err != nil {
		t.Error(err)
	}
	// check fields that are not in schema
	for key, prop := range dbt.Properties {
		if _, ok := s[key]; !ok {
			js, err := json.MarshalIndent(&prop, " ", " ")
			if err != nil {
				js = []byte{}
			}
			t.Errorf("%s property '%s' is not defined in schema\n%s", dbType, key, string(js))
		}
	}
	// check removed fields from schema
	for key := range s {
		if _, ok := dbt.Properties[key]; !ok {
			t.Errorf("%s schema field '%s' is no longer supported", dbType, key)
		}
	}
}
