package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceOpenSearch() *schema.Resource {
	return &schema.Resource{
		Description:   serviceDescription("OpenSearch"),
		CreateContext: resourceOpenSearchCreate,
		ReadContext:   resourceOpenSearchRead,
		UpdateContext: resourceOpenSearchUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(upcloud.ManagedDatabaseServiceTypeOpenSearch),
			schemaOpenSearchEngine(),
			schemaOpenSearchAccessControl(),
		),
	}
}

func resourceOpenSearchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := d.Set("type", string(upcloud.ManagedDatabaseServiceTypeOpenSearch)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDatabaseCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	if d.HasChanges("access_control", "extended_access_control") {
		client := meta.(*service.Service)
		aclReq := request.ModifyManagedDatabaseAccessControlRequest{
			ServiceUUID:         d.Id(),
			ACLsEnabled:         upcloud.BoolPtr(d.Get("access_control").(bool)),
			ExtendedACLsEnabled: upcloud.BoolPtr(d.Get("extended_access_control").(bool)),
		}
		_, err := client.ModifyManagedDatabaseAccessControl(ctx, &aclReq)
		if err != nil {
			return utils.HandleResourceError(d.Get("name").(string), d, err)
		}
	}

	return resourceOpenSearchRead(ctx, d, meta)
}

func resourceOpenSearchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceDatabaseRead(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	client := meta.(*service.Service)
	aclReq := request.GetManagedDatabaseAccessControlRequest{ServiceUUID: d.Id()}
	acl, err := client.GetManagedDatabaseAccessControl(ctx, &aclReq)
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}
	if err = d.Set("access_control", acl.ACLsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("extended_access_control", acl.ExtendedACLsEnabled); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceOpenSearchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChanges("access_control", "extended_access_control") {
		client := meta.(*service.Service)
		aclReq := request.ModifyManagedDatabaseAccessControlRequest{
			ServiceUUID:         d.Id(),
			ACLsEnabled:         upcloud.BoolPtr(d.Get("access_control").(bool)),
			ExtendedACLsEnabled: upcloud.BoolPtr(d.Get("extended_access_control").(bool)),
		}
		_, err := client.ModifyManagedDatabaseAccessControl(ctx, &aclReq)
		if err != nil {
			return utils.HandleResourceError(d.Get("name").(string), d, err)
		}
	}

	diags := resourceDatabaseUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	diags = append(diags, resourceOpenSearchRead(ctx, d, meta)...)
	return diags
}

func schemaOpenSearchEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"properties": {
			Description: "Database Engine properties for OpenSearch",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: properties.GetSchemaMap(upcloud.ManagedDatabaseServiceTypeOpenSearch),
			},
		},
	}
}

func schemaOpenSearchAccessControl() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"access_control": {
			Type:        schema.TypeBool,
			Description: "Enables users access control for OpenSearch service. User access control rules will only be enforced if this attribute is enabled.",
			Optional:    true,
			Computed:    true,
		},
		"extended_access_control": {
			Type:        schema.TypeBool,
			Description: "Grant access to top-level `_mget`, `_msearch` and `_bulk` APIs. Users are limited to perform operations on indices based on the user-specific access control rules.",
			Optional:    true,
			Computed:    true,
		},
	}
}
