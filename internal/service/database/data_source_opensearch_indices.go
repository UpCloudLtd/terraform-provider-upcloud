package database

import (
	"context"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceOpenSearchIndices() *schema.Resource {
	return &schema.Resource{
		Description: "OpenSearch indices",
		ReadContext: dataSourceOpenSearchIndicesRead,
		Schema: map[string]*schema.Schema{
			"indices": {
				Description: "Available indices for OpenSearch",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        schemaOpenSearchIndex(),
			},
			"service": {
				Description: "Service's UUID for which these indices belongs to",
				Type:        schema.TypeString,
				Required:    true,
				Computed:    false,
			},
		},
	}
}

func schemaOpenSearchIndex() *schema.Resource {
	return &schema.Resource{
		Description: "OpenSearch index",
		Schema: map[string]*schema.Schema{
			"create_time": {
				Description: "Timestamp indicating the creation time of the index.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"docs": {
				Description: "Number of documents stored in the index.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"health": {
				Description: "Health status of the index e.g. `green`, `yellow`, or `red`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"index_name": {
				Description: "Name of the index.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"number_of_replicas": {
				Description: "Number of replicas configured for the index.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"number_of_shards": {
				Description: "Number of shards configured & used by the index.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"read_only_allow_delete": {
				Description: "Indicates whether the index is in a read-only state that permits deletion of the entire index. This attribute can be automatically set to true in certain scenarios where the node disk space exceeds the flood stage.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"size": {
				Description: "Size of the index in bytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"status": {
				Description: "Status of the index e.g. `open` or `closed`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceOpenSearchIndicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)
	serviceID := d.Get("service").(string)

	indices, err := client.GetManagedDatabaseIndices(ctx, &request.GetManagedDatabaseIndicesRequest{
		ServiceUUID: serviceID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceID)

	if err := d.Set("service", serviceID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("indices", buildOpenSearchIndices(indices)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func buildOpenSearchIndices(indices []upcloud.ManagedDatabaseIndex) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0)
	for _, index := range indices {
		maps = append(maps, map[string]interface{}{
			"create_time":            index.CreateTime.UTC().Format(time.RFC3339Nano),
			"docs":                   index.Docs,
			"health":                 index.Health,
			"index_name":             index.IndexName,
			"status":                 index.Status,
			"number_of_replicas":     index.NumberOfReplicas,
			"number_of_shards":       index.NumberOfShards,
			"read_only_allow_delete": index.ReadOnlyAllowDelete,
			"size":                   index.Size,
		})
	}

	return maps
}
