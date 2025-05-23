package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
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
	if os.Getenv("TF_ACC_DB_PROPS") == "" {
		t.Skip("Skipping database properties validation, because we have CI workflow for updating these automatically. Set TF_ACC_DB_PROPS environment variable to run these tests.")
	}

	dbTypes := []upcloud.ManagedDatabaseServiceType{
		upcloud.ManagedDatabaseServiceTypeMySQL,
		upcloud.ManagedDatabaseServiceTypeOpenSearch,
		upcloud.ManagedDatabaseServiceTypePostgreSQL,
		upcloud.ManagedDatabaseServiceTypeRedis, //nolint:staticcheck // To be removed when Redis support has been removed
		upcloud.ManagedDatabaseServiceTypeValkey,
	}

	for _, dbType := range dbTypes {
		t.Run(string(dbType), func(t *testing.T) {
			s := properties.GetSchemaMap(dbType)
			testProperties(t, string(dbType), s)
		})
	}
}

func testProperties(t *testing.T, dbType string, s map[string]*schema.Schema) {
	svc := utils.NewServiceWithCredentialsFromEnv(t)
	dbt, err := svc.GetManagedDatabaseServiceType(context.Background(), &request.GetManagedDatabaseServiceTypeRequest{
		Type: dbType,
	})
	if err != nil {
		t.Error(err)
	}
	// check fields that are not in schema
	for key := range dbt.Properties {
		if _, ok := s[key]; !ok {
			t.Errorf("%s property '%s' is not defined in schema. Run `make generate` to update properties.", dbType, key)
		}
	}
	// check removed fields from schema
	for key := range s {
		if _, ok := dbt.Properties[key]; !ok {
			t.Errorf("%s schema field '%s' is no longer supported. Run `make generate` to update properties.", dbType, key)
		}
	}
}
