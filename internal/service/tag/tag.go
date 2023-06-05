package tag

import (
	"context"
	"regexp"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceTag() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is deprecated, use tags schema in server resource",
		CreateContext: resourceTagCreate,
		ReadContext:   resourceTagRead,
		UpdateContext: resourceTagUpdate,
		DeleteContext: resourceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Description:  "Free form text representing the meaning of the tag",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"name": {
				Description: "The value representing the tag",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.Any(validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile("[a-zA-Z0-9_]"), "")),
			},
			"servers": {
				Description: "A collection of servers that have been assigned the tag",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	createTagRequest := &request.CreateTagRequest{
		Tag: upcloud.Tag{
			Name: d.Get("name").(string),
		},
	}
	if description, ok := d.GetOk("description"); ok {
		createTagRequest.Description = description.(string)
	}
	if servers, ok := d.GetOk("servers"); ok {
		servers := servers.(*schema.Set)
		serversList := make([]string, len(servers.List()))
		for i := range serversList {
			serversList[i] = servers.List()[i].(string)
		}

		createTagRequest.Servers = serversList
	}

	tag, err := client.CreateTag(ctx, createTagRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tag.Name)

	return resourceTagRead(ctx, d, meta)
}

func resourceTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	tags, err := client.GetTags(ctx)
	if err != nil {
		diag.FromErr(err)
	}

	tagID := d.Id()
	var tag *upcloud.Tag

	for _, value := range tags.Tags {
		if value.Name == tagID {
			tag = &value
			break
		}
	}

	if tag == nil {
		return diag.Errorf("Unable to locate tag named %s", tagID)
	}

	if err := d.Set("name", tag.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", tag.Description); err != nil {
		return diag.FromErr(err)
	}

	servers := []string{}
	for _, server := range tag.Servers {
		servers = append(servers, server)
	}

	if err := d.Set("servers", servers); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	r := &request.ModifyTagRequest{
		Name: d.Id(),
	}

	r.Tag.Name = d.Id()
	if d.HasChange("description") {
		_, newDescription := d.GetChange("description")
		r.Tag.Description = newDescription.(string)
	}
	if d.HasChange("servers") {
		_, newServers := d.GetChange("servers")

		servers := newServers.(*schema.Set)
		serversList := make([]string, len(servers.List()))
		for i := range serversList {
			serversList[i] = servers.List()[i].(string)
		}
		r.Tag.Servers = serversList
	}

	_, err := client.ModifyTag(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTagRead(ctx, d, meta)
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	deleteTagRequest := &request.DeleteTagRequest{
		Name: d.Id(),
	}
	err := client.DeleteTag(ctx, deleteTagRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
