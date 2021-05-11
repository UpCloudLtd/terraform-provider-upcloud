package upcloud

import (
	"context"
	"net/url"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const bucketKey = "bucket"
const numRetries = 5

func resourceUpCloudObjectStorage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectStorageCreate,
		ReadContext:   resourceObjectStorageRead,
		UpdateContext: resourceObjectStorageUpdate,
		DeleteContext: resourceObjectStorageDelete,
		Schema: map[string]*schema.Schema{
			"size": {
				Description:  "The size of the object storage instance in gigabytes",
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
				Description: "The zone in which the object storage instance will be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the object storage instance to be created",
				Required:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description:  "The description of the object storage instance to be created",
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
			bucketKey: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Description:  "The name of the bucket",
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 127),
						},
					},
				},
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

	objStorage, err := createObjectStorage(client, &req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create object storage",
			Detail:   err.Error(),
		})
		return diags
	}

	if v, ok := d.GetOk(bucketKey); ok {
		conn, err := getBucketConnection(objStorage.URL, req.AccessKey, req.SecretKey)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, bucketDetails := range v.([]interface{}) {
			details := bucketDetails.(map[string]interface{})

			err = conn.MakeBucket(ctx, details["name"].(string), minio.MakeBucketOptions{})

			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(objStorage.UUID)

	return copyObjectStorageDetails(objStorage, d)
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

	return copyObjectStorageDetails(objectDetails, d)
}

func resourceObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*service.Service)

	if d.HasChanges([]string{"size", "access_key", "secret_key", "description"}...) {
		req := request.ModifyObjectStorageRequest{UUID: d.Id()}

		req.Size = d.Get("size").(int)
		req.AccessKey = d.Get("access_key").(string)
		req.SecretKey = d.Get("secret_key").(string)
		req.Description = d.Get("description").(string)

		_, err := modifyObjectStorage(client, &req)
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

	if d.HasChange(bucketKey) {
		conn, err := getBucketConnection(
			d.Get("url").(string),
			d.Get("access_key").(string),
			d.Get("secret_key").(string),
		)

		if err != nil {
			return diag.FromErr(err)
		}

		bucketsToDelete, bucketsToAdd := getNewAndDeletedBucketNames(d)

		for _, bucket := range bucketsToDelete {
			err = conn.RemoveBucket(ctx, bucket)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		for _, bucket := range bucketsToAdd {
			err = conn.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return diag.FromErr(err)
			}
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

func copyObjectStorageDetails(objectDetails *upcloud.ObjectStorageDetails, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set("name", objectDetails.Name)
	_ = d.Set("url", objectDetails.URL)
	_ = d.Set("description", objectDetails.Description)
	_ = d.Set("size", objectDetails.Size)
	_ = d.Set("state", objectDetails.State)
	_ = d.Set("created", objectDetails.Created)
	_ = d.Set("zone", objectDetails.Zone)
	_ = d.Set("used_space", objectDetails.UsedSpace)

	buckets, err := getBuckets(objectDetails, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set(bucketKey, buckets)

	return diag.Diagnostics{}
}

func getBucketConnection(URL, accessKey, secretKey string) (*minio.Client, error) {
	urlObj, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	return minio.New(urlObj.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
}

func getNewAndDeletedBucketNames(d *schema.ResourceData) ([]string, []string) {
	beforeNames := make([]string, 0)
	afterNames := make([]string, 0)

	before, after := d.GetChange(bucketKey)

	for _, item := range before.([]interface{}) {
		valueMap := item.(map[string]interface{})
		beforeNames = append(beforeNames, valueMap["name"].(string))
	}

	for _, item := range after.([]interface{}) {
		valueMap := item.(map[string]interface{})
		afterNames = append(afterNames, valueMap["name"].(string))
	}

	return getMissing(beforeNames, afterNames), getMissing(afterNames, beforeNames)
}

func getMissing(expected, found []string) []string {
	var missing []string
	for _, expectedName := range expected {
		nameFound := false
		for _, foundName := range found {
			if foundName == expectedName {
				nameFound = true
				break
			}
		}

		if !nameFound {
			if missing == nil {
				missing = make([]string, 0, 1)
			}

			missing = append(missing, expectedName)
		}
	}

	return missing
}

func getBuckets(objectDetails *upcloud.ObjectStorageDetails, d *schema.ResourceData) ([]map[string]interface{}, error) {
	conn, err := getBucketConnection(
		objectDetails.URL,
		d.Get("access_key").(string),
		d.Get("secret_key").(string),
	)

	if err != nil {
		return nil, err
	}

	// sometimes fails here because the buckets aren't redy yet
	var bucketInfo []minio.BucketInfo
	for trys := 0; trys < numRetries; trys++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		bucketInfo, err = conn.ListBuckets(ctx)
		cancel()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	bucketNames := make([]map[string]interface{}, 0, len(bucketInfo))
	for _, bucket := range bucketInfo {
		bucketNames = append(bucketNames, map[string]interface{}{"name": bucket.Name})
	}

	return bucketNames, nil
}

func createObjectStorage(client *service.Service, req *request.CreateObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	var (
		err        error
		objStorage *upcloud.ObjectStorageDetails
	)
	for try := 0; try < numRetries; try++ {
		// calls to the function seem to fail occasionally, so call it in a retry loop
		objStorage, err = client.CreateObjectStorage(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	return objStorage, err
}

func modifyObjectStorage(client *service.Service, req *request.ModifyObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	var (
		err        error
		objStorage *upcloud.ObjectStorageDetails
	)
	for try := 0; try < numRetries; try++ {
		objStorage, err = client.ModifyObjectStorage(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	return objStorage, err
}
