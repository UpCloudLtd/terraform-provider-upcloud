package objectstorage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	bucketKey             string = "bucket"
	accessKeyEnvVarPrefix string = "UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_"
	secretKeyEnvVarPrefix string = "UPCLOUD_OBJECT_STORAGE_SECRET_KEY_"
	numRetries            int    = 5
	accessKeyMinLength    int    = 4
	accessKeyMaxLength    int    = 255
	secretKeyMinLength    int    = 8
	secretKeyMaxLength    int    = 255

	deprecationMessage string = "The `upcloud_object_storage` resource manages previous generatation object storage instances that will reach end of life (EOL) by the end of 2024. For new instances, consider using the new Object Storage product managed with `upcloud_managed_object_storage` resource."
)

type objectStorageKeyType string

func ResourceObjectStorage() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		Description: fmt.Sprintf(`~> %s

This resource represents an UpCloud Object Storage instance, which provides S3 compatible storage.`, deprecationMessage),
		DeprecationMessage: deprecationMessage,
		CreateContext:      resourceObjectStorageCreate,
		ReadContext:        resourceObjectStorageRead,
		UpdateContext:      resourceObjectStorageUpdate,
		DeleteContext:      resourceObjectStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"size": {
				Description:  "The size of the object storage instance in gigabytes",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{250, 500, 1000}),
			},
			"access_key": {
				Description: `The access key used to identify user.
				Can be set to an empty string, which will tell the provider to get the access key from environment variable.
				The environment variable should be "UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_{name}".
				{name} is the name given to object storage instance (so not the resource label), it should be all uppercased
				and all dashes (-) should be replaced with underscores (_). For example, object storage named "my-files" would
				use environment variable named "UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_MY_FILES".`,
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: createKeyValidationFunc("access_key", accessKeyMinLength, accessKeyMaxLength),
			},
			"secret_key": {
				Description: `The secret key used to authenticate user.
				Can be set to an empty string, which will tell the provider to get the secret key from environment variable.
				The environment variable should be "UPCLOUD_OBJECT_STORAGE_SECRET_KEY_{name}".
				{name} is the name given to object storage instance (so not the resource label), it should be all uppercased
				and all dashes (-) should be replaced with underscores (_). For example, object storage named "my-files" would
				use environment variable named "UPCLOUD_OBJECT_STORAGE_SECRET_KEY_MY_FILES".`,
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: createKeyValidationFunc("secret_key", secretKeyMinLength, secretKeyMaxLength),
			},
			"zone": {
				Description: "The zone in which the object storage instance will be created, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					EnableLegacyTypeSystemApplyErrors: true,
					EnableLegacyTypeSystemPlanErrors:  true,
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

	accessKey, _, err := getAccessKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	secretKey, _, err := getSecretKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Size = d.Get("size").(int)
	req.Zone = d.Get("zone").(string)
	req.Name = d.Get("name").(string)
	req.AccessKey = accessKey
	req.SecretKey = secretKey
	req.Description = d.Get("description").(string)

	objStorage, err := createObjectStorage(ctx, client, &req)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	if v, ok := d.GetOk(bucketKey); ok {
		conn, err := GetBucketConnection(objStorage.URL, req.AccessKey, req.SecretKey)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, bucketDetails := range v.(*schema.Set).List() {
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

	objectDetails, err := client.GetObjectStorageDetails(ctx, &request.GetObjectStorageDetailsRequest{
		UUID: uuid,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	return copyObjectStorageDetails(objectDetails, d)
}

func resourceObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*service.Service)

	accessKey, _, err := getAccessKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	secretKey, _, err := getSecretKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges([]string{"size", "access_key", "secret_key", "description"}...) {
		req := request.ModifyObjectStorageRequest{UUID: d.Id()}

		req.Size = d.Get("size").(int)
		req.AccessKey = accessKey
		req.SecretKey = secretKey
		req.Description = d.Get("description").(string)

		_, err := modifyObjectStorage(ctx, client, &req)
		if err != nil {
			var diags diag.Diagnostics
			diags = append(diags, diag.FromErr(err)...)
			return diags
		}
	}

	if d.HasChange(bucketKey) {
		conn, err := GetBucketConnection(
			d.Get("url").(string),
			accessKey,
			secretKey,
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

	err := client.DeleteObjectStorage(ctx, &request.DeleteObjectStorageRequest{
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

	accessKey, accessKeyFromEnv, err := getAccessKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	secretKey, secretKeyFromEnv, err := getSecretKey(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if accessKeyFromEnv {
		_ = d.Set("access_key", "")
	}

	if secretKeyFromEnv {
		_ = d.Set("secret_key", "")
	}

	buckets, err := getBuckets(objectDetails.URL, accessKey, secretKey)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set(bucketKey, buckets)

	return diag.Diagnostics{}
}

func appendRetryError(title string, wrapper error, err error) error {
	if wrapper == nil {
		return fmt.Errorf("%s:\n- %w", title, err)
	}

	return fmt.Errorf("%s\n- %w", wrapper.Error(), err)
}

func createObjectStorage(ctx context.Context, client *service.Service, req *request.CreateObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	var (
		allErrors  error
		err        error
		objStorage *upcloud.ObjectStorageDetails
	)
	for try := 0; try < numRetries; try++ {
		// calls to the function seem to fail occasionally, so call it in a retry loop
		objStorage, err = client.CreateObjectStorage(ctx, req)
		if err == nil {
			break
		}
		allErrors = appendRetryError("unable to create object storage", allErrors, err)
		time.Sleep(time.Second)
	}
	return objStorage, allErrors
}

func modifyObjectStorage(ctx context.Context, client *service.Service, req *request.ModifyObjectStorageRequest) (*upcloud.ObjectStorageDetails, error) {
	var (
		allErrors  error
		err        error
		objStorage *upcloud.ObjectStorageDetails
	)
	for try := 0; try < numRetries; try++ {
		objStorage, err = client.ModifyObjectStorage(ctx, req)
		if err == nil {
			break
		}
		allErrors = appendRetryError("unable to modify object storage", allErrors, err)
		time.Sleep(time.Second)
	}
	return objStorage, allErrors
}

func getBuckets(URL, accessKey, secretKey string) ([]map[string]interface{}, error) {
	conn, err := GetBucketConnection(URL, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	// sometimes fails here because the buckets aren't ready yet
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

func GetBucketConnection(URL, accessKey, secretKey string) (*minio.Client, error) {
	urlObj, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	return minio.New(urlObj.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
}
