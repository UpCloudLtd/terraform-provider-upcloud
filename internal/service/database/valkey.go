package database

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceValkey() *schema.Resource {
	return &schema.Resource{
		Description:   serviceDescription("Valkey"),
		CreateContext: resourceValkeyCreate,
		ReadContext:   resourceValkeyRead,
		UpdateContext: resourceValkeyUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(upcloud.ManagedDatabaseServiceTypeValkey),
			schemaValkeyEngine(),
		),
	}
}

func resourceValkeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := d.Set("type", string(upcloud.ManagedDatabaseServiceTypeValkey)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDatabaseCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	return resourceValkeyRead(ctx, d, meta)
}

func resourceValkeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatabaseRead(ctx, d, meta)
}

func resourceValkeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceDatabaseUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	diags = append(diags, resourceValkeyRead(ctx, d, meta)...)
	return diags
}

func schemaValkeyEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"properties": {
			Description: "Database Engine properties for Valkey",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: properties.GetSchemaMap(upcloud.ManagedDatabaseServiceTypeValkey),
			},
		},
	}
}
