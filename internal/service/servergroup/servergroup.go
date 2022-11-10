package servergroup

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const serverGroupNotFoundErrorCode = "GROUP_NOT_FOUND"

func ResourceServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerGroupCreate,
		ReadContext:   resourceServerGroupRead,
		// UpdateContext: resourceServerGroupUpdate,
		DeleteContext: resourceServerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "Title of your server group",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"labels": {
				Description: "Labels for your server group",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
			"members": {
				Description: "UUIDs of the servers that are members of this group",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
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
				for example via API, UpCloud Control Panel or upctl (UpCloud CLI)`,
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	req := &request.CreateServerGroupRequest{
		Title: d.Get("title").(string),
	}

	if d.Get("anti_affinity").(bool) {
		req.AntiAffinity = upcloud.True
	} else {
		req.AntiAffinity = upcloud.False
	}

	members, membersWereSet := d.GetOk("members")
	if membersWereSet {
		membersSlice, err := utils.SetOfStringsToSlice(ctx, members)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Creating server group failed",
				Detail:   err.Error(),
			})

			return diags
		}

		req.Members = membersSlice
	}

	labels, labelsWereSet := d.GetOk("labels")
	if labelsWereSet {
		labelsSlice, err := utils.MapOfStringsToLabelSlice(ctx, labels)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Creating server group failed",
				Detail:   err.Error(),
			})

			return diags
		}

		req.Labels = &labelsSlice
	}

	group, err := svc.CreateServerGroup(ctx, req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Creating server group failed",
			Detail:   err.Error(),
		})

		return diags
	}

	d.SetId(group.UUID)

	diags = append(diags, resourceServerGroupRead(ctx, d, meta)...)
	return diags
}

func resourceServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	group, err := svc.GetServerGroup(ctx, &request.GetServerGroupRequest{UUID: d.Id()})
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == serverGroupNotFoundErrorCode {
			diags = append(diags, utils.DiagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}

		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	d.Set("title", group.Title)
	d.Set("anti_affinity", group.AntiAffinity.Bool())
	d.Set("members", group.Members)

	labels := map[string]string{}
	for _, label := range group.Labels {
		labels[label.Key] = label.Value
	}

	d.Set("labels", labels)

	return diags
}

// func resourceServerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return resourceServerGroupRead(ctx, d, meta)
// }

func resourceServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	err := svc.DeleteServerGroup(ctx, &request.DeleteServerGroupRequest{UUID: d.Id()})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
