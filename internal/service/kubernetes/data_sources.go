package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Returns Kubernetes cluter.",
		ReadContext: dataSourceClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster": {
				Description: "ID of the Kubernetes cluster.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"config": {
				Description: "Kubernetes config",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return diags
}
