package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   serviceDescription("MySQL"),
		CreateContext: resourceMySQLCreate,
		ReadContext:   resourceMySQLRead,
		UpdateContext: resourceMySQLUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(),
			schemaMySQLEngine(),
		),
	}
}

func resourceMySQLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := d.Set("type", string(upcloud.ManagedDatabaseServiceTypeMySQL)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDatabaseCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	return resourceMySQLRead(ctx, d, meta)
}

func resourceMySQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatabaseRead(ctx, d, meta)
}

func resourceMySQLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceDatabaseUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	diags = append(diags, resourceMySQLRead(ctx, d, meta)...)
	return diags
}

func schemaMySQLEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"properties": {
			Description: "Database Engine properties for MySQL",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: properties.GetSchemaMap(upcloud.ManagedDatabaseServiceTypeMySQL),
			},
		},
	}
}
