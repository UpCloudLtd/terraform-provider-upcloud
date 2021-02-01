package upcloud

import (
	"context"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudObjectStorage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectStorageCreate,
		ReadContext:   resourceObjectStorageRead,
		UpdateContext: resourceObjectStorageUpdate,
		DeleteContext: resourceObjectStorageDelete,
		Schema: map[string]*schema.Schema{
			"size": {
				Description:  "The size of the object storage bucket in gigabytes",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{250, 500, 1000}),
			},
			"access_key": {
				Description:  "The access key used to identify user",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(4, 255),
			},
			"secret_key": {
				Description:  "The secret key used to authenticate user",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(8, 255),
			},
			"zone": {
				Description: "The zone in which the object storage bucket will be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the object storage bucket to be created",
				Required:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description:  "The description of the object storage bucket to be created",
				Required:     true,
				Type:         schema.TypeString,
				DefaultFunc:  func() (interface{}, error) { return "managed by terraform", nil },
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"used_space": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceObjectStorageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		diags diag.Diagnostics
		req   request.CreateObjectStorageRequest
	)

	client := m.(*service.Service)

	req.Size = d.Get("size").(int)
	req.Zone = d.Get("zone").(string)
	req.Name = d.Get("name").(string)
	req.AccessKey = d.Get("access_key").(string)
	req.SecretKey = d.Get("secret_key").(string)
	req.Description = d.Get("description").(string)

	objStorage, err := client.CreateObjectStorage(&req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create object storage",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(objStorage.UUID)

	copyObjectStorageDetails(objStorage, d)

	return diags
}

func resourceObjectStorageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*service.Service)

	uuid := d.Id()

	objectDetails, err := client.GetObjectStorageDetails(&request.GetObjectStorageDetailsRequest{
		UUID: uuid,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	copyObjectStorageDetails(objectDetails, d)

	return diag.Diagnostics{}
}

func resourceObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*service.Service)

	if d.HasChanges([]string{"size", "access_key", "secret_key", "description"}...) {

		req := request.ModifyObjectStorageRequest{UUID: d.Id()}

		req.Size = d.Get("size").(int)
		req.AccessKey = d.Get("access_key").(string)
		req.SecretKey = d.Get("secret_key").(string)
		req.Description = d.Get("description").(string)

		_, err := client.ModifyObjectStorage(&req)
		if err != nil {
			var diags diag.Diagnostics
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to modify object storage",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	return resourceObjectStorageRead(ctx, d, m)
}

func resourceObjectStorageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*service.Service)

	var diags diag.Diagnostics

	err := client.DeleteObjectStorage(&request.DeleteObjectStorageRequest{
		UUID: d.Id(),
	})

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete object storage",
			Detail:   err.Error(),
		})

	} else {
		d.SetId("")
	}

	return diags
}

func copyObjectStorageDetails(objectDetails *upcloud.ObjectStorageDetails, d *schema.ResourceData) {
	d.Set("name", objectDetails.Name)
	d.Set("url", objectDetails.URL)
	d.Set("description", objectDetails.Description)
	d.Set("size", objectDetails.Size)
	d.Set("state", objectDetails.State)
	d.Set("created", objectDetails.Created)
	d.Set("zone", objectDetails.Zone)
	d.Set("used_space", objectDetails.UsedSpace)
}
