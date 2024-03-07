package database

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a user in managed database",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: schemaUser(),
	}
}

func schemaUser() map[string]*schema.Schema {
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
		"authentication": {
			Description: "MySQL only, authentication type.",
			Type:        schema.TypeString,
			// caching_sha2_password is used by default, but that can not be set via Default field as that would set the value also for non MySQL users.
			Optional: true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
				string(upcloud.ManagedDatabaseUserAuthenticationCachingSHA2Password),
				string(upcloud.ManagedDatabaseUserAuthenticationMySQLNativePassword),
			}, false)),
		},
		"pg_access_control": {
			Description:   "PostgreSQL access control object.",
			ConflictsWith: []string{"redis_access_control", "opensearch_access_control"},
			Type:          schema.TypeList,
			Optional:      true,
			MaxItems:      1,
			Elem: &schema.Resource{
				Schema: schemaPostgreSQLUserAccessControl(),
			},
		},
		"redis_access_control": {
			Description: "Redis access control object.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: schemaRedisUserAccessControl(),
			},
		},
		"opensearch_access_control": {
			Description: "OpenSearch access control object.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: schemaOpenSearchUserAccessControl(),
			},
		},
	}
}

func schemaPostgreSQLUserAccessControl() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"allow_replication": {
			Description: "Grant replication privilege",
			Type:        schema.TypeBool,
			Default:     true,
			Optional:    true,
		},
	}
}

func schemaRedisUserAccessControl() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"categories": {
			Description: "Set access control to all commands in specified categories.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"channels": {
			Description: "Set access control to Pub/Sub channels.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"commands": {
			Description: "Set access control to commands.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"keys": {
			Description: "Set access control to keys.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func schemaOpenSearchUserAccessControl() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"rules": {
			Description: "Set user access control rules.",
			Type:        schema.TypeList,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Description: "Set index name, pattern or top level API.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"permission": {
						Description: "Set permission access.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
							string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionAdmin),
							string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionDeny),
							string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionRead),
							string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionReadWrite),
							string(upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermissionWrite),
						}, false)),
					},
				},
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	if d.HasChange("type") && d.Get("type").(string) != string(upcloud.ManagedDatabaseUserTypeNormal) {
		return diag.FromErr(fmt.Errorf("only type `normal` users can be created"))
	}

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
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
	req := &request.CreateManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
	}
	switch serviceDetails.Type {
	case upcloud.ManagedDatabaseServiceTypeMySQL:
		if val, ok := d.Get("authentication").(string); ok && val != "" {
			req.Authentication = upcloud.ManagedDatabaseUserAuthenticationType(val)
		} else {
			req.Authentication = upcloud.ManagedDatabaseUserAuthenticationCachingSHA2Password
		}
	case upcloud.ManagedDatabaseServiceTypePostgreSQL:
		if v, ok := d.Get("pg_access_control.0.allow_replication").(bool); ok {
			req.PGAccessControl = &upcloud.ManagedDatabaseUserPGAccessControl{
				AllowReplication: &v,
			}
		}
	case upcloud.ManagedDatabaseServiceTypeRedis:
		req.RedisAccessControl = redisAccessControlFromResourceData(d)
	case upcloud.ManagedDatabaseServiceTypeOpenSearch:
		req.OpenSearchAccessControl = openSearchAccessControlFromResourceData(d)
	}

	if _, err = client.CreateManagedDatabaseUser(ctx, req); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.MarshalID(serviceID, d.Get("username").(string)))

	tflog.Info(ctx, "managed database user created", map[string]interface{}{
		"service_name": serviceDetails.Name, "username": d.Get("username").(string), "service_uuid": serviceID,
	})

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var serviceID, username string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &username); err != nil {
		return diag.FromErr(err)
	}

	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return utils.HandleResourceError(d.Get("username").(string), d, err)
	}

	// If service UUID is not set already set it based on the Id. This is the case for example when importing existing user.
	if _, ok := d.GetOk("service"); !ok {
		err := d.Set("service", serviceID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	userDetails, err := client.GetManagedDatabaseUser(ctx, &request.GetManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("username").(string), d, err)
	}

	tflog.Info(ctx, "managed database user read", map[string]interface{}{
		"service_name": serviceDetails.Name, "username": username, "service_uuid": serviceID,
	})
	return copyUserDetailsToResource(d, userDetails)
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot modify a user while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	var username string
	if utils.UnmarshalID(d.Id(), &serviceID, &username) != nil {
		return diag.FromErr(err)
	}
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &request.ModifyManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username,
		Password:    d.Get("password").(string),
	}
	if serviceDetails.Type == upcloud.ManagedDatabaseServiceTypeMySQL {
		if val, ok := d.Get("authentication").(string); ok && val != "" {
			req.Authentication = upcloud.ManagedDatabaseUserAuthenticationType(val)
		}
	}
	if _, err = client.ModifyManagedDatabaseUser(ctx, req); err != nil {
		return diag.FromErr(err)
	}

	switch serviceDetails.Type {
	case upcloud.ManagedDatabaseServiceTypePostgreSQL:
		if d.HasChange("pg_access_control.0") {
			if _, err := modifyPostgreSQLUserAccessControl(ctx, client, d); err != nil {
				return diag.FromErr(err)
			}
		}
	case upcloud.ManagedDatabaseServiceTypeRedis:
		if d.HasChange("redis_access_control.0") {
			if _, err := modifyRedisUserAccessControl(ctx, client, d); err != nil {
				return diag.FromErr(err)
			}
		}
	case upcloud.ManagedDatabaseServiceTypeOpenSearch:
		if d.HasChange("opensearch_access_control.0") {
			if _, err := modifyOpenSearchUserAccessControl(ctx, client, d); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	tflog.Info(ctx, "managed database user updated", map[string]interface{}{
		"service_name": serviceDetails.Name, "username": username, "service_uuid": serviceID,
	})

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	if d.Get("type").(string) == string(upcloud.ManagedDatabaseUserTypePrimary) {
		if d.HasChange("username") {
			return diag.FromErr(fmt.Errorf("primary username cannot be changed %q", d.Id()))
		}
		tflog.Debug(ctx, "ignoring delete for primary user %q", map[string]interface{}{"uuid": d.Id()})
		return nil
	}

	serviceID := d.Get("service").(string)
	serviceDetails, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}
	if !serviceDetails.Powered {
		return diag.FromErr(fmt.Errorf("cannot delete a user while managed database %v (%v) is powered off", serviceDetails.Name, serviceID))
	}

	var username string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &username); err != nil {
		return diag.FromErr(err)
	}
	serviceDetails, err = resourceUpCloudManagedDatabaseWaitState(ctx, serviceID, meta,
		d.Timeout(schema.TimeoutCreate), resourceUpcloudManagedDatabaseModifiableStates...)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteManagedDatabaseUser(ctx, &request.DeleteManagedDatabaseUserRequest{
		ServiceUUID: serviceID,
		Username:    username,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "managed database user deleted", map[string]interface{}{
		"service_name": serviceDetails.Name, "username": username, "service_uuid": serviceID,
	})

	return nil
}

func copyUserDetailsToResource(d *schema.ResourceData, details *upcloud.ManagedDatabaseUser) diag.Diagnostics {
	if err := d.Set("username", details.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", details.Password); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", details.Type); err != nil {
		return diag.FromErr(err)
	}
	if details.Authentication != "" {
		if err := d.Set("authentication", details.Authentication); err != nil {
			return diag.FromErr(err)
		}
	}
	if details.PGAccessControl != nil {
		if err := d.Set("pg_access_control", []map[string]interface{}{
			{
				"allow_replication": details.PGAccessControl.AllowReplication,
			},
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if details.RedisAccessControl != nil {
		if err := d.Set("redis_access_control", []map[string][]string{
			{
				"categories": *details.RedisAccessControl.Categories,
				"channels":   *details.RedisAccessControl.Channels,
				"commands":   *details.RedisAccessControl.Commands,
				"keys":       *details.RedisAccessControl.Keys,
			},
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if details.OpenSearchAccessControl != nil {
		rules := make([]map[string]interface{}, 0)
		for _, rule := range *details.OpenSearchAccessControl.Rules {
			rules = append(rules, map[string]interface{}{
				"index":      rule.Index,
				"permission": string(rule.Permission),
			})
		}

		if err := d.Set("opensearch_access_control", []map[string]interface{}{
			{
				"rules": rules,
			},
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func modifyPostgreSQLUserAccessControl(ctx context.Context, svc *service.Service, d *schema.ResourceData) (*upcloud.ManagedDatabaseUser, error) {
	req := &request.ModifyManagedDatabaseUserAccessControlRequest{
		ServiceUUID: d.Get("service").(string),
		Username:    d.Get("username").(string),
		PGAccessControl: &upcloud.ManagedDatabaseUserPGAccessControl{
			AllowReplication: upcloud.BoolPtr(d.Get("pg_access_control.0.allow_replication").(bool)),
		},
	}
	return svc.ModifyManagedDatabaseUserAccessControl(ctx, req)
}

func modifyRedisUserAccessControl(ctx context.Context, svc *service.Service, d *schema.ResourceData) (*upcloud.ManagedDatabaseUser, error) {
	req := &request.ModifyManagedDatabaseUserAccessControlRequest{
		ServiceUUID:        d.Get("service").(string),
		Username:           d.Get("username").(string),
		RedisAccessControl: redisAccessControlFromResourceData(d),
	}
	return svc.ModifyManagedDatabaseUserAccessControl(ctx, req)
}

func redisAccessControlFromResourceData(d *schema.ResourceData) *upcloud.ManagedDatabaseUserRedisAccessControl {
	acl := &upcloud.ManagedDatabaseUserRedisAccessControl{}
	if v, ok := d.Get("redis_access_control.0.categories").([]interface{}); ok {
		categories := make([]string, len(v))
		for i := range v {
			categories[i] = v[i].(string)
		}
		acl.Categories = &categories
	}
	if v, ok := d.Get("redis_access_control.0.channels").([]interface{}); ok {
		channels := make([]string, len(v))
		for i := range v {
			channels[i] = v[i].(string)
		}
		acl.Channels = &channels
	}
	if v, ok := d.Get("redis_access_control.0.commands").([]interface{}); ok {
		commands := make([]string, len(v))
		for i := range v {
			commands[i] = v[i].(string)
		}
		acl.Commands = &commands
	}
	if v, ok := d.Get("redis_access_control.0.keys").([]interface{}); ok {
		keys := make([]string, len(v))
		for i := range v {
			keys[i] = v[i].(string)
		}
		acl.Keys = &keys
	}
	return acl
}

func openSearchAccessControlFromResourceData(d *schema.ResourceData) *upcloud.ManagedDatabaseUserOpenSearchAccessControl {
	acl := &upcloud.ManagedDatabaseUserOpenSearchAccessControl{}
	if v, ok := d.Get("opensearch_access_control.0.rules").([]interface{}); ok {
		rules := make([]upcloud.ManagedDatabaseUserOpenSearchAccessControlRule, len(v))
		for i := range v {
			if index, ok := d.Get(fmt.Sprintf("opensearch_access_control.0.rules.%d.index", i)).(string); ok {
				rules[i].Index = index
			}
			if permission, ok := d.Get(fmt.Sprintf("opensearch_access_control.0.rules.%d.permission", i)).(string); ok {
				rules[i].Permission = upcloud.ManagedDatabaseUserOpenSearchAccessControlRulePermission(permission)
			}
		}
		acl.Rules = &rules
	}
	return acl
}

func modifyOpenSearchUserAccessControl(ctx context.Context, svc *service.Service, d *schema.ResourceData) (*upcloud.ManagedDatabaseUser, error) {
	req := &request.ModifyManagedDatabaseUserAccessControlRequest{
		ServiceUUID:             d.Get("service").(string),
		Username:                d.Get("username").(string),
		OpenSearchAccessControl: openSearchAccessControlFromResourceData(d),
	}
	return svc.ModifyManagedDatabaseUserAccessControl(ctx, req)
}
