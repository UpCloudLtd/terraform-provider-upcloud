package database

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestIsManagedDatabaseFullyCreated(t *testing.T) {
	db := &upcloud.ManagedDatabase{
		Backups: make([]upcloud.ManagedDatabaseBackup, 0),
		State:   upcloud.ManagedDatabaseStatePoweroff,
		Users:   make([]upcloud.ManagedDatabaseUser, 0),
	}
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.State = upcloud.ManagedDatabaseStateRunning
	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Users = append(db.Users, upcloud.ManagedDatabaseUser{})
	db.Backups = make([]upcloud.ManagedDatabaseBackup, 0)
	if isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want false got true %+v", db)
	}

	db.Backups = append(db.Backups, upcloud.ManagedDatabaseBackup{})
	if !isManagedDatabaseFullyCreated(db) {
		t.Errorf("isManagedDatabaseFullyCreated failed want true got false %+v", db)
	}
}

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
		schemaDatabaseCommonProperties(),
		schemaPostgreSQLProperties(),
	)
	testProperties(t, "pg", s)
}

func TestMySQLProperties(t *testing.T) {
	s := utils.JoinSchemas(
		schemaDatabaseCommonProperties(),
		schemaMySQLProperties(),
	)
	testProperties(t, "mysql", s)
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
	for key, prop := range dbt.Properties {
		if _, ok := s[key]; !ok {
			js, err := json.MarshalIndent(&prop, " ", " ")
			if err != nil {
				js = []byte{}
			}
			t.Logf("%s property '%s' is not defined in schema\n%s", dbType, key, string(js))
		}
	}
}
