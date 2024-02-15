package tag

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceTags() *schema.Resource {
	return &schema.Resource{
		Description: "Data-source is deprecated.",
		ReadContext: dataSourceTagsRead,
		Schema: map[string]*schema.Schema{
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Description: "Free form text representing the meaning of the tag",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The value representing the tag",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"servers": {
							Description: "A collection of servers that have been assigned the tag",
							Type:        schema.TypeSet,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	tags, err := client.GetTags(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching tags: %s", err))
	}

	var values []map[string]interface{}

	for _, tag := range tags.Tags {
		servers := []string{}
		for _, server := range tag.Servers {
			servers = append(servers, server)
		}

		value := map[string]interface{}{
			"name":        tag.Name,
			"description": tag.Description,
			"servers":     servers,
		}

		values = append(values, value)
	}

	if err := d.Set("tags", values); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())

	return diags
}
