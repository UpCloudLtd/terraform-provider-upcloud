package kubernetes

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNodeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer service",
		CreateContext: resourceNodeGroupCreate,
		ReadContext:   resourceNodeGroupRead,
		UpdateContext: resourceNodeGroupUpdate,
		DeleteContext: resourceNodeGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cluster": {
				Description: "ID of the Kubernetes cluster.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the service must be unique within customer account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"plan": {
				Description: "The pricing plan used for the node group",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"labels": {
				Description: "Labels contain key-value pairs to classify the node group",
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceNodeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return diags
}

func resourceNodeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	cluster, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: d.Id()})
	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setClusterResourceData(d, cluster); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceNodeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return diags
}

func resourceNodeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func setClusterResourceData(d *schema.ResourceData, c *upcloud.KubernetesCluster) (diags diag.Diagnostics) {
	if err := d.Set("name", c.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", c.Zone); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
