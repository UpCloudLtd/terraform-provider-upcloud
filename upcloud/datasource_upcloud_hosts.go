package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUpCloudHosts() *schema.Resource {
	return &schema.Resource{
		Description: `Returns a list of available UpCloud hosts. 
		A host identifies the host server that virtual machines are run on. 
		Only hosts on private cloud to which the calling account has access to are available through this resource.`,
		ReadContext: dataSourceUpCloudHostsRead,
		Schema: map[string]*schema.Schema{
			"hosts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_id": {
							Type:        schema.TypeInt,
							Description: "The unique id of the host",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Free form text describing the host",
							Computed:    true,
						},
						"zone": {
							Type:        schema.TypeString,
							Description: "The zone the host is in",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUpCloudHostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	hosts, err := client.GetHosts()

	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching hosts: %s", err))
	}

	var values []map[string]interface{}

	for _, host := range hosts.Hosts {
		value := map[string]interface{}{
			"host_id":     host.ID,
			"description": host.Description,
			"zone":        host.Zone,
		}

		values = append(values, value)
	}

	if err := d.Set("hosts", values); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())

	return diags
}
