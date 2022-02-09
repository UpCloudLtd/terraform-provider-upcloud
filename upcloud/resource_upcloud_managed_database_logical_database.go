package upcloud

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudManagedDatabaseLogicalDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a logical database in managed database",
		CreateContext: resourceUpCloudManagedDatabaseLogicalDatabaseCreate,
		ReadContext:   resourceUpCloudManagedDatabaseLogicalDatabaseRead,
		DeleteContext: resourceUpCloudManagedDatabaseLogicalDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				serviceID, name := splitManagedDatabaseSubResourceID(data.Id())
				if serviceID == "" || name == "" {
					return nil, fmt.Errorf("invalid import id. Format: <managedDatabaseUUID>/<logicalDatabaseName>")
				}
				if err := data.Set("service", serviceID); err != nil {
					return nil, err
				}
				if err := data.Set("name", name); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{data}, nil
			},
		},
		Schema: schemaUpCloudManagedDatabaseLogicalDatabase(),
	}
}

func schemaUpCloudManagedDatabaseLogicalDatabase() map[string]*schema.Schema {
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

func resourceUpCloudManagedDatabaseLogicalDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
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
	_, err = client.CreateManagedDatabaseLogicalDatabase(&request.CreateManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: serviceID,
		Name:        d.Get("name").(string),
		LCCType:     d.Get("character_set").(string),
		LCCollate:   d.Get("collation").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildManagedDatabaseSubResourceID(serviceID, d.Get("name").(string)))
	log.Printf("[INFO] managed database logical database %v/%v (%v/%v) created",
		serviceDetails.Name, d.Get("name").(string),
		serviceID, d.Get("name").(string))

	return resourceUpCloudManagedDatabaseLogicalDatabaseRead(ctx, d, meta)
}

func resourceUpCloudManagedDatabaseLogicalDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID, name := splitManagedDatabaseSubResourceID(d.Id())

	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudDatabaseNotFoundErrorCode {
			var diags diag.Diagnostics
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	ldbs, err := client.GetManagedDatabaseLogicalDatabases(&request.GetManagedDatabaseLogicalDatabasesRequest{
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
		var diags diag.Diagnostics
		diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(
			&upcloud.Error{
				ErrorCode:    upcloudLogicalDatabaseNotFoundErrorCode,
				ErrorMessage: fmt.Sprintf("logical database %q was not found", name),
			},
			d.Get("name").(string)))
		d.SetId("")
		return diags
	}

	log.Printf("[DEBUG] managed database logical database %v/%v (%v/%v) read",
		serviceDetails.Name, name,
		serviceID, name)
	return copyManagedDatabaseLogicalDatabaseDetailsToResource(d, details)
}

func resourceUpCloudManagedDatabaseLogicalDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot delete a logical database while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	serviceID, name := splitManagedDatabaseSubResourceID(d.Id())
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteManagedDatabaseLogicalDatabase(&request.DeleteManagedDatabaseLogicalDatabaseRequest{
		ServiceUUID: serviceID,
		Name:        name,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] managed database logical database %v/%v (%v/%v) deleted",
		serviceDetails.Name, name,
		serviceID, name)
	return nil
}

func copyManagedDatabaseLogicalDatabaseDetailsToResource(d *schema.ResourceData, details *upcloud.ManagedDatabaseLogicalDatabase) diag.Diagnostics {
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
