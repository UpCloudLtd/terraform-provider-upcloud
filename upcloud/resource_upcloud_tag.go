package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

func resourceUpCloudTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudTagCreate,
		ReadContext:   resourceUpCloudTagRead,
		UpdateContext: resourceUpCloudTagUpdate,
		DeleteContext: resourceUpCloudTagDelete,
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

func resourceUpCloudTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tag, err := client.CreateTag(createTagRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tag.Name)

	return resourceUpCloudTagRead(ctx, d, meta)
}

func resourceUpCloudTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*service.Service)

	var diags diag.Diagnostics

	tags, err := client.GetTags()

	if err != nil {
		diag.FromErr(err)
	}

	tagId := d.Id()
	var tag *upcloud.Tag

	for _, value := range tags.Tags {

		if value.Name == tagId {
			tag = &value
			break
		}
	}

	if tag == nil {
		return diag.Errorf("Unable to locate tag named %s", tagId)
	}

	if err := d.Set("name", tag.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", tag.Description); err != nil {
		return diag.FromErr(err)
	}

	var servers = []string{}
	for _, server := range tag.Servers {
		servers = append(servers, server)
	}

	if err := d.Set("servers", servers); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUpCloudTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_, err := client.ModifyTag(r)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudTagRead(ctx, d, meta)
}

func resourceUpCloudTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	deleteTagRequest := &request.DeleteTagRequest{
		Name: d.Id(),
	}
	err := client.DeleteTag(deleteTagRequest)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
