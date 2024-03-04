package database

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
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

func TestPostgreSQLProperties(t *testing.T) {
	s := utils.JoinSchemas(
		schemaRDBMSDatabaseCommonProperties(),
		schemaDatabaseCommonProperties(),
		schemaPostgreSQLProperties(),
	)
	testProperties(t, "pg", s)
}

func TestMySQLProperties(t *testing.T) {
	s := utils.JoinSchemas(
		schemaRDBMSDatabaseCommonProperties(),
		schemaDatabaseCommonProperties(),
		schemaMySQLProperties(),
	)
	testProperties(t, "mysql", s)
}

func TestRedisProperties(t *testing.T) {
	s := utils.JoinSchemas(
		schemaDatabaseCommonProperties(),
		schemaRedisProperties(),
	)
	testProperties(t, "redis", s)
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
			t.Logf("%s property '%s' is not defined in schema\n%s", dbType, key, string(js))
		}
	}
	// check removed fields from schema
	for key := range s {
		if _, ok := dbt.Properties[key]; !ok {
			t.Logf("%s schema field '%s' is no longer supported", dbType, key)
		}
	}
}
