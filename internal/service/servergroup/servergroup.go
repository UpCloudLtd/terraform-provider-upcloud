package servergroup

import (
	"context"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerGroupCreate,
		ReadContext:   resourceServerGroupRead,
		UpdateContext: resourceServerGroupUpdate,
		DeleteContext: resourceServerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "Title of your server group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"labels": {
				Description: "Labels for your server group",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"members": {
				Description: "UUIDs of the servers that are members of this group",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"anti_affinity": {
				Description: `Is group an anti-affinity group. Setting this to true will result in
				all servers in the group being placed on separate compute hosts.

				NOTE: this is an experimental feature. The anti-affinity policy is "best-effort" and it is not
				guaranteed that all the servers will end up on a separate compute hosts. You can verify if the
				anti-affinity policies are met by requesting a server group details from API. For more information
				please see UpCloud API documentation on server groups
				
				NOTE: anti-affinity policies are only applied on server start. This means that if anti-affinity
				policies in server group are not met, you need to manually restart the servers in said group,
				for example via API or UpCloud Control Panel`,
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return resourceServerGroupRead(ctx, d, meta)
}

func resourceServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return diags
}

func resourceServerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceServerGroupRead(ctx, d, meta)
}

func resourceServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	err := svc.DeleteServerGroup(ctx, &request.DeleteServerGroupRequest{UUID: d.Id()})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
