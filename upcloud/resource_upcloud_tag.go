package upcloud

import (
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUpCloudTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpCloudTagCreate,
		Update: resourceUpCloudTagUpdate,
		Delete: resourceUpCloudTagDelete,
		Read:   resourceUpCloudTagRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"servers": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUpCloudTagRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceUpCloudTagCreate(d *schema.ResourceData, meta interface{}) error {
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
		servers := servers.([]interface{})
		serversList := make([]string, len(servers))
		for i := range serversList {
			serversList[i] = servers[i].(string)
		}
		createTagRequest.Servers = serversList
	}

	tag, err := client.CreateTag(createTagRequest)

	if err != nil {
		return err
	}

	d.SetId(tag.Name)

	return nil
}

func resourceUpCloudTagUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)
	r := &request.ModifyTagRequest{}

	if d.HasChange("description") {
		_, newDescription := d.GetChange("description")
		r.Description = newDescription.(string)
	}
	if d.HasChange("servers") {
		_, newServers := d.GetChange("servers")
		r.Servers = newServers.([]string)
	}

	_, err := client.ModifyTag(r)

	if err != nil {
		return err
	}

	return nil
}

func resourceUpCloudTagDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	deleteTagRequest := &request.DeleteTagRequest{
		Name: d.Id(),
	}
	err := client.DeleteTag(deleteTagRequest)

	if err != nil {
		return err
	}

	return nil
}
