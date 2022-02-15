package upcloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const bucketKey = "bucket"
const numRetries = 5
const AccessKeyEnvVarPrefix = "UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_"
const SecretKeyEnvVarPrefix = "UPCLOUD_OBJECT_STORAGE_SECRET_KEY_"
const accessKeyMinLength = 4
const accessKeyMaxLength = 255
const secretKeyMinLength = 8
const secretKeyMaxLength = 255

func resourceUpCloudObjectStorage() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Object Storage instance, which provides S3 compatible storage.",
		CreateContext: resourceObjectStorageCreate,
		ReadContext:   resourceObjectStorageRead,
		UpdateContext: resourceObjectStorageUpdate,
		DeleteContext: resourceObjectStorageDelete,
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
				Description:      "The access key used to identify user",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: createKeyValidationFunc("access_key", accessKeyMinLength, accessKeyMaxLength),
			},
			"secret_key": {
				Description:      "The secret key used to authenticate user",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: createKeyValidationFunc("secret_key", secretKeyMinLength, secretKeyMaxLength),
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
				Type:     schema.TypeSet,
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

	accessKey, _, err := getAccessKey(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Access key not found",
			Detail:   err.Error(),
		})
		return diags
	}

	secretKey, _, err := getSecretKey(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Secret key not found",
			Detail:   err.Error(),
		})
		return diags
	}

	req.Size = d.Get("size").(int)
	req.Zone = d.Get("zone").(string)
	req.Name = d.Get("name").(string)
	req.AccessKey = accessKey
	req.SecretKey = secretKey
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

	objectDetails, err := client.GetObjectStorageDetails(&request.GetObjectStorageDetailsRequest{
		UUID: uuid,
	})

	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudObjectStorageNotFoundErrorCode {
			var diags diag.Diagnostics
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}
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

	log.Println(fmt.Sprintf("\033[32m[INFO] State: %+v\033[0m", d.State()))

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

	for _, item := range before.(*schema.Set).List() {
		valueMap := item.(map[string]interface{})
		beforeNames = append(beforeNames, valueMap["name"].(string))
	}

	for _, item := range after.(*schema.Set).List() {
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

func getBuckets(URL, accessKey, secretKey string) ([]map[string]interface{}, error) {
	conn, err := getBucketConnection(URL, accessKey, secretKey)

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

// Attempts to get access key.
// Second return value is a bool, set to true if key value was retrived from env variable
func getAccessKey(d *schema.ResourceData) (string, bool, error) {
	configVal := d.Get("access_key").(string)

	// If config value is set to something else then empty string, just use it
	if configVal != "" {
		return configVal, false, nil
	}

	// If config value is empty string, use environment variable
	objectStorageName := d.Get("name").(string)
	envVarKey := generateObjectStorageEnvVarKey(AccessKeyEnvVarPrefix, objectStorageName)
	envVarValue, envVarSet := os.LookupEnv(envVarKey)

	if !envVarSet {
		return "", false, fmt.Errorf("access_key config field for object storage %s is set to empty string and environment variable %s is not set", objectStorageName, envVarKey)
	}

	length := len(envVarValue)

	if length < accessKeyMinLength {
		return "", false, fmt.Errorf("access_key set in environment variable %s is too short; minimum length is %d, got %d", envVarKey, accessKeyMinLength, length)
	}

	if length > accessKeyMaxLength {
		return "", false, fmt.Errorf("access_key set in environment variable %s is too long; maximum length is %d, got %d", envVarKey, accessKeyMaxLength, length)
	}

	return envVarValue, true, nil
}

// Attempts to get secret key.
// Second return value is a bool, set to true if key value was revtrived from env variable
func getSecretKey(d *schema.ResourceData) (string, bool, error) {
	configVal := d.Get("secret_key").(string)

	// If config value is set to something else then empty string, just use it
	if configVal != "" {
		return configVal, false, nil
	}

	// If config value is empty string, use environment variable
	objectStorageName := d.Get("name").(string)
	envVarKey := generateObjectStorageEnvVarKey(SecretKeyEnvVarPrefix, objectStorageName)
	envVarValue, envVarSet := os.LookupEnv(envVarKey)

	if !envVarSet {
		return "", false, fmt.Errorf("secret_key config field for object storage %s is set to empty string and environment variable %s is not set", objectStorageName, envVarKey)
	}

	length := len(envVarValue)

	if length < secretKeyMinLength {
		return "", false, fmt.Errorf("secret_key set in environment variable %s is too short; minimum length is %d, got %d", envVarKey, secretKeyMinLength, length)
	}

	if length > secretKeyMaxLength {
		return "", false, fmt.Errorf("secret_key set in environment variable %s is too long; maximum length is %d, got %d", envVarKey, secretKeyMaxLength, length)
	}

	return envVarValue, true, nil
}

func generateObjectStorageEnvVarKey(prefix, objectStorageName string) string {
	name := strings.ToUpper(strings.Replace(objectStorageName, "-", "_", -1))
	return fmt.Sprintf("%s%s", prefix, name)
}

type objectStorageKeyType string

const (
	objectStorageKeyTypeAccess objectStorageKeyType = "access_key"
	objectStorageKeyTypeSecret objectStorageKeyType = "secret_key"
)

func createKeyValidationFunc(attrName objectStorageKeyType, minLength, maxLength int) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		key := val.(string)

		// For access and secret keys empty string means that they should be taken from env vars
		if key == "" {
			var envVarPrefix string

			switch attrName {
			case objectStorageKeyTypeAccess:
				envVarPrefix = AccessKeyEnvVarPrefix
			case objectStorageKeyTypeSecret:
				envVarPrefix = SecretKeyEnvVarPrefix
			default:
				return diag.Errorf("unknown attribute name for creating object storage keys validation function: %s; this is a provider error", attrName)
			}

			if !utils.EnvKeyExists(envVarPrefix) {
				return diag.Errorf("%s set to empty string, but no environment variables for it found", attrName)
			}

			return diag.Diagnostics{}
		}

		length := len(key)

		if length < minLength {
			return diag.Errorf("%s too short; minimum length is %d, got %d", attrName, minLength, length)
		}

		if length > maxLength {
			return diag.Errorf("%s too long; max length is %d, got %d", attrName, maxLength, length)
		}

		return diag.Diagnostics{}
	}
}
