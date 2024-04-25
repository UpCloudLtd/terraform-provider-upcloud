package database

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLogicalDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a logical database in managed database",
		CreateContext: resourceLogicalDatabaseCreate,
		ReadContext:   resourceLogicalDatabaseRead,
		DeleteContext: resourceLogicalDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: schemaLogicalDatabase(),
	}
}

func schemaLogicalDatabase() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service": {
			Description: "Service's UUID for which this user belongs to",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "Name of the logical database",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"character_set": {
			Description:      "Default character set for the database (LC_CTYPE)",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validateManagedDatabaseLocale),
		},
		"collation": {
			Description:      "Default collation for the database (LC_COLLATE)",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validateManagedDatabaseLocale),
		},
	}
}

func resourceLogicalDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot create a logical database while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	if d.HasChanges("character_set", "collation") && serviceDetails.Type != upcloud.ManagedDatabaseServiceTypePostgreSQL {
		return diag.FromErr(fmt.Errorf("setting character_set or collation is only possible for PostgreSQL service"))
	}

	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.CreateManagedDatabaseLogicalDatabase(ctx, &request.CreateManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: serviceID,
		Name:        d.Get("name").(string),
		LCCType:     d.Get("character_set").(string),
		LCCollate:   d.Get("collation").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, d.Get("name").(string)))

	tflog.Info(ctx, "managed database logical database created", map[string]interface{}{
		"service_name": serviceDetails.Name, "name": d.Get("name").(string), "service_uuid": serviceID,
	})

	return resourceLogicalDatabaseRead(ctx, d, meta)
}

func resourceLogicalDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var serviceID, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}

	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	// If service UUID is not set already set it based on the Id. This is the case for example when importing existing user.
	if _, ok := d.GetOk("service"); !ok {
		err := d.Set("service", serviceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	ldbs, err := client.GetManagedDatabaseLogicalDatabases(ctx, &request.GetManagedDatabaseLogicalDatabasesRequest{
		ServiceUUID: serviceID,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	var details *upcloud.ManagedDatabaseLogicalDatabase
	for i, ldb := range ldbs {
		if ldb.Name == name {
			details = &ldbs[i]
		}
	}
	if details == nil {
		// We need to manually construct the error here as the actual `err` value is not relevant here
		return utils.HandleResourceError(d.Get("name").(string), d, &upcloud.Problem{Status: http.StatusNotFound})
	}

	tflog.Info(ctx, "managed database logical database read", map[string]interface{}{
		"service_name": serviceDetails.Name, "name": name, "service_uuid": serviceID,
	})

	return copyLogicalDatabaseDetailsToResource(d, details)
}

func resourceLogicalDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot delete a logical database while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	var name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &name); err != nil {
		return diag.FromErr(err)
	}
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteManagedDatabaseLogicalDatabase(ctx, &request.DeleteManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "managed database logical database deleted", map[string]interface{}{
		"service_name": serviceDetails.Name, "name": name, "service_uuid": serviceID,
	})

	return nil
}

func copyLogicalDatabaseDetailsToResource(d *schema.ResourceData, details *upcloud.ManagedDatabaseLogicalDatabase) diag.Diagnostics {
	setFields := []struct {
		name string
		val  interface{}
	}{
		{name: "name", val: details.Name},
		{name: "character_set", val: details.LCCType},
		{name: "collation", val: details.LCCollate},
	}

	for _, sf := range setFields {
		if err := d.Set(sf.name, sf.val); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

var validateManagedDatabaseLocale = validation.StringMatch(
	regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}\.[A-z0-9-]+$`),
	"invalid locale; must be in form en_US.UTF8 (language_TERRITORY.CODEPOINT)",
)
