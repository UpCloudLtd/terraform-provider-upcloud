package upcloud

import (
	"context"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func dataSourceUpCloudTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUpCloudTagsRead,
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

func dataSourceUpCloudTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	tags, err := client.GetTags()

	if err != nil {
		return diag.FromErr(fmt.Errorf("Error fetching Tags: %s", err))
	}

	var values []map[string]interface{}

	for _, tag := range tags.Tags {

		var servers = []string{}
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
