package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description:   serviceDescription("PostgreSQL"),
		CreateContext: resourcePostgreSQLCreate,
		ReadContext:   resourcePostgreSQLRead,
		UpdateContext: resourcePostgreSQLUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(upcloud.ManagedDatabaseServiceTypePostgreSQL),
			schemaPostgreSQLEngine(),
		),
	}
}

func resourcePostgreSQLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := d.Set("type", string(upcloud.ManagedDatabaseServiceTypePostgreSQL)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDatabaseCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	return resourcePostgreSQLRead(ctx, d, meta)
}

func resourcePostgreSQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatabaseRead(ctx, d, meta)
}

func resourcePostgreSQLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceDatabaseUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	client := meta.(*service.Service)

	if !d.HasChange("powered") {
		if d.HasChange("properties.0.version") {
			diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
			if diags.HasError() {
				return diags
			}
		}
	} else {
		switch d.Get("powered").(bool) {
		// Power off
		case false:
			if d.HasChange("properties.0.version") {
				diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
				if diags.HasError() {
					return diags
				}
			}
			diags = append(diags, resourceDatabasePoweredUpdate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
		// Power on
		case true:
			diags = append(diags, resourceDatabasePoweredUpdate(ctx, d, meta)...)
			if diags.HasError() {
				return diags
			}
			if d.HasChange("properties.0.version") {
				diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
				if diags.HasError() {
					return diags
				}
			}
		}
	}

	diags = append(diags, resourcePostgreSQLRead(ctx, d, meta)...)
	return diags
}

func schemaPostgreSQLEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sslmode": {
			Description: "SSL Connection Mode for PostgreSQL",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"properties": {
			Description: "Database Engine properties for PostgreSQL",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: properties.GetSchemaMap(upcloud.ManagedDatabaseServiceTypePostgreSQL),
			},
		},
	}
}
