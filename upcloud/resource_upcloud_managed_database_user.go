package upcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudManagedDatabaseUser() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a user in managed database",
		CreateContext: resourceUpCloudManagedDatabaseUserCreate,
		ReadContext:   resourceUpCloudManagedDatabaseUserRead,
		UpdateContext: resourceUpCloudManagedDatabaseUserUpdate,
		DeleteContext: resourceUpCloudManagedDatabaseUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
				serviceID, user := splitManagedDatabaseSubResourceID(data.Id())
				if serviceID == "" || user == "" {
					return nil, fmt.Errorf("invalid import id. Format: <managedDatabaseUUID>/<username>")
				}
				if err := data.Set("service", serviceID); err != nil {
					return nil, err
				}
				if err := data.Set("username", user); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{data}, nil
			},
		},
		Schema: schemaUpCloudManagedDatabaseUser(),
	}
}

func schemaUpCloudManagedDatabaseUser() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service": {
			Description: "Service's UUID for which this user belongs to",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"username": {
			Description: "Name of the database user",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Description:      "Password for the database user. Defaults to a random value",
			Type:             schema.TypeString,
			Sensitive:        true,
			Computed:         true,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(8, 256)),
		},
		"type": {
			Description: "Type of the user. Only normal type users can be created",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceUpCloudManagedDatabaseUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	if d.HasChange("type") && d.Get("type").(string) != string(upcloud.ManagedDatabaseUserTypeNormal) {
		return diag.FromErr(fmt.Errorf("only type `normal` users can be created"))
	}

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot create a user while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.CreateManagedDatabaseUser(&request.CreateManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildManagedDatabaseSubResourceID(serviceID, d.Get("username").(string)))
	log.Printf("[INFO] managed database user %v/%v (%v/%v) created",
		serviceDetails.Name, d.Get("username").(string),
		serviceID, d.Get("username").(string))

	return resourceUpCloudManagedDatabaseUserRead(ctx, d, meta)
}

func resourceUpCloudManagedDatabaseUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID, username := splitManagedDatabaseSubResourceID(d.Id())

	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudDatabaseNotFoundErrorCode {
			var diags diag.Diagnostics
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("username").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	userDetails, err := client.GetManagedDatabaseUser(&request.GetManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username})
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudDatabaseUserNotFoundErrorCode {
			var diags diag.Diagnostics
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("username").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] managed database user %v/%v (%v/%v) read",
		serviceDetails.Name, username,
		serviceID, username)
	return copyManagedDatabaseUserDetailsToResource(d, userDetails)
}

func resourceUpCloudManagedDatabaseUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot modify a user while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	serviceID, username := splitManagedDatabaseSubResourceID(d.Id())
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ModifyManagedDatabaseUser(&request.ModifyManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username,
		Password:    d.Get("password").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] managed database user %v/%v (%v/%v) deleted",
		serviceDetails.Name, username,
		serviceID, username)
	return resourceUpCloudManagedDatabaseUserRead(ctx, d, meta)
}

func resourceUpCloudManagedDatabaseUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	if d.Get("type").(string) == string(upcloud.ManagedDatabaseUserTypePrimary) {
		if d.HasChange("username") {
			return diag.FromErr(fmt.Errorf("primary username cannot be changed %q", d.Id()))
		}
		log.Printf("[DEBUG] ignoring delete for primary user %q", d.Id())
		return nil
	}

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot delete a user while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	serviceID, username := splitManagedDatabaseSubResourceID(d.Id())
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteManagedDatabaseUser(&request.DeleteManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] managed database user %v/%v (%v/%v) deleted",
		serviceDetails.Name, username,
		serviceID, username)
	return nil
}

func copyManagedDatabaseUserDetailsToResource(d *schema.ResourceData, details *upcloud.ManagedDatabaseUser) diag.Diagnostics {
	setFields := []struct {
		name string
		val  interface{}
	}{
		{name: "username", val: details.Username},
		{name: "password", val: details.Password},
		{name: "type", val: string(details.Type)},
	}

	for _, sf := range setFields {
		if err := d.Set(sf.name, sf.val); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
