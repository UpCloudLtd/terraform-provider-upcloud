package upcloud

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/objectstorage"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/minio/minio-go/v7"
)

const (
	objectStorageTestExpectedDescription = "My object storage"
	objectStorageTestExpectedZone        = "pl-waw1"
	objectStorageTestExpectedKey         = "an access key"
	objectStorageTestExpectedSecret      = "a secret key"
	objectStorageTestRunPrefix           = "testacc-"
	accessKeyEnvVarPrefix                = "UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_"
	secretKeyEnvVarPrefix                = "UPCLOUD_OBJECT_STORAGE_SECRET_KEY_"
)

var (
	objectStorageTestRunID         = time.Now().Unix()
	objectStorageTestExpectedName1 = fmt.Sprintf("%s%d-1", objectStorageTestRunPrefix, objectStorageTestRunID)
	objectStorageTestExpectedName2 = fmt.Sprintf("%s%d-2", objectStorageTestRunPrefix, objectStorageTestRunID)
	objectStorageTestExpectedName3 = fmt.Sprintf("%s%d-3", objectStorageTestRunPrefix, objectStorageTestRunID)
	objectStorageTestExpectedName4 = fmt.Sprintf("%s%d-4", objectStorageTestRunPrefix, objectStorageTestRunID)
)

func init() {
	resource.AddTestSweepers("object_storage_cleanup", &resource.Sweeper{
		Name: "object_storage_cleanup",
		F: func(_ string) error {
			username, ok := os.LookupEnv("UPCLOUD_USERNAME")
			if !ok {
				return fmt.Errorf("UPCLOUD_USERNAME must be set for acceptance tests")
			}

			password, ok := os.LookupEnv("UPCLOUD_PASSWORD")
			if !ok {
				return fmt.Errorf("UPCLOUD_PASSWORD must be set for acceptance tests")
			}

			client := retryablehttp.NewClient()

			requestTimeout := 120 * time.Second

			service := newUpCloudServiceConnection(username, password, client.HTTPClient, requestTimeout)

			objectStorages, err := service.GetObjectStorages(context.Background())
			if err != nil {
				return err
			}

			for _, objectStorage := range objectStorages.ObjectStorages {
				if !strings.HasPrefix(objectStorage.Name, objectStorageTestRunPrefix) {
					continue
				}

				err = service.DeleteObjectStorage(
					context.Background(),
					&request.DeleteObjectStorageRequest{
						UUID: objectStorage.UUID,
					})
				if err != nil {
					return err
				}
			}

			return nil
		},
	})
}

// TestMain is boilerplate needed for -sweep command line parameters to work
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestUpCloudObjectStorage_basic(t *testing.T) {
	const expectedSize = "250"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		CheckDestroy:             verifyObjectStorageDoesNotExist(objectStorageTestExpectedName1),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedSize, objectStorageTestExpectedName1, objectStorageTestExpectedDescription, objectStorageTestExpectedZone, objectStorageTestExpectedKey, objectStorageTestExpectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "name"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "description"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", objectStorageTestExpectedName1),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", objectStorageTestExpectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", objectStorageTestExpectedZone),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName1),
				),
			},
		},
	})
}

func TestUpCloudObjectStorage_basic_update(t *testing.T) {
	const expectedSize = "500"

	const expectedUpdatedSize = "1000"
	const expectedUpdatedDescription = "My Updated data collection"
	const expectedUpdatedKey = "an updated access key"
	const expectedUpdatedSecret = "an updated secret"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		CheckDestroy:             verifyObjectStorageDoesNotExist(objectStorageTestExpectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedDescription, objectStorageTestExpectedZone, objectStorageTestExpectedKey, objectStorageTestExpectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", objectStorageTestExpectedName2),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", objectStorageTestExpectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", objectStorageTestExpectedZone),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName2),
				),
			},
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedUpdatedSize,
					objectStorageTestExpectedName2,
					expectedUpdatedDescription,
					objectStorageTestExpectedZone,
					expectedUpdatedKey,
					expectedUpdatedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedUpdatedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedUpdatedDescription),
					verifyObjectStorageExists(expectedUpdatedKey, expectedUpdatedSecret, objectStorageTestExpectedName2),
				),
			},
		},
	})
}

func TestUpCloudObjectStorage_default_values(t *testing.T) {
	const expectedSize = "500"
	const expectedUpdatedSize = "1000"
	const expectedUpdatedKey = "an updated access key"
	const expectedUpdatedSecret = "an updated secret"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		CheckDestroy:             verifyObjectStorageDoesNotExist(objectStorageTestExpectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceDefaultsConfig(
					expectedSize, objectStorageTestExpectedName3, objectStorageTestExpectedZone, objectStorageTestExpectedKey, objectStorageTestExpectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", objectStorageTestExpectedZone),
				),
			},
			{
				Config: testUpCloudObjectStorageInstanceDefaultsConfig(
					expectedUpdatedSize,
					objectStorageTestExpectedName3,
					objectStorageTestExpectedZone,
					expectedUpdatedKey,
					expectedUpdatedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedUpdatedSize),
				),
			},
		},
	})
}

func TestUpCloudObjectStorage_bucket_management(t *testing.T) {
	const expectedSize = "500"
	const expectedBucketName1 = "bucket1"
	const expectedBucketName2 = "bucket2"
	const expectedBucketName3 = "bucket3"
	const expectedBucketName4 = "bucket4"
	const expectedBucketName5 = "bucket5"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		CheckDestroy:             verifyObjectStorageDoesNotExist(objectStorageTestExpectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedZone,
					objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1, expectedBucketName2,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", objectStorageTestExpectedName2),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", objectStorageTestExpectedZone),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName2),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName2),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedZone,
					objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName2, expectedBucketName1,
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedZone,
					objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName2),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1),
					verifyBucketDoesNotExist(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName2),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedZone,
					objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1, expectedBucketName3, expectedBucketName4,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName2),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1),
					verifyBucketDoesNotExist(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName2),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName3),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName4),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, objectStorageTestExpectedName2, objectStorageTestExpectedZone,
					objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName4, expectedBucketName5,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, objectStorageTestExpectedName2),
					verifyBucketDoesNotExist(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName1),
					verifyBucketDoesNotExist(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName2),
					verifyBucketDoesNotExist(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName3),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName4),
					verifyBucketExists(objectStorageTestExpectedKey, objectStorageTestExpectedSecret, expectedBucketName5),
				),
			},
		},
	})
}

// We bundle creating object storage using env vars and import because import relies on passing access and secret key as env vars
func TestUpCloudObjectStorage_keys_env_vars_and_import(t *testing.T) {
	name := objectStorageTestExpectedName4
	zone := "pl-waw1"
	desc := "just some random stuff"
	size := "250"
	accessKey := objectStorageTestExpectedKey
	secretKey := objectStorageTestExpectedSecret

	updatedAccessKey := "someaccesskeymodified"
	updatedSecretKey := "somesupersecretyoyo"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			name := strings.ToUpper(strings.Replace(name, "-", "_", -1))
			accessKeyEnvVarName := fmt.Sprintf("%s%s", accessKeyEnvVarPrefix, name)
			secretKeyEnvVarName := fmt.Sprintf("%s%s", secretKeyEnvVarPrefix, name)

			testAccPreCheck(t)
			os.Setenv(accessKeyEnvVarName, accessKey)
			os.Setenv(secretKeyEnvVarName, secretKey)
		},
		ProtoV5ProviderFactories: testAccProviderFactories,
		CheckDestroy:             verifyObjectStorageDoesNotExist(name),
		Steps: []resource.TestStep{
			{
				// Pass empty strings as access and secret keys to check if those values will be taken from env vars
				Config: testUpCloudObjectStorageInstanceConfig(size, name, desc, zone, "", ""),
				// Just check if object storage was actually created
				Check: resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "name", name),
			},
			{
				// Check if you can update keys by just putting them into config
				Config: testUpCloudObjectStorageInstanceConfig(size, name, desc, zone, updatedAccessKey, updatedSecretKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "access_key", updatedAccessKey),
					resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "secret_key", updatedSecretKey),
				),
			},
			{
				// Check if you can modify buckets with updated keys
				Config: testUpCloudObjectStorageInstanceConfigWithSingleBucket(size, name, desc, zone, updatedAccessKey, updatedSecretKey),
				Check:  resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "bucket.0.name", "test"),
			},
			{
				// Remove keys from config again to check if you can switch back to using env vars
				// Also check if removing buckets will work in one go here
				Config: testUpCloudObjectStorageInstanceConfig(size, name, desc, zone, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "access_key", ""),
					resource.TestCheckResourceAttr("upcloud_object_storage.my_storage", "secret_key", ""),
					resource.TestCheckNoResourceAttr("upcloud_object_storage.my_storage", "bucket.0.name"),
				),
			},
			{
				// Verify import
				ResourceName:      "upcloud_object_storage.my_storage",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testUpCloudObjectStorageInstanceConfig(size, name, description, zone, accessKey, secretKey string) string {
	return fmt.Sprintf(`
		resource "upcloud_object_storage" "my_storage" {
			size  = %s
			name  = "%s"
			description = "%s"
			zone  = "%s"
			access_key = "%s"
			secret_key = "%s"
		}
`, size, name, description, zone, accessKey, secretKey)
}

func testUpCloudObjectStorageInstanceConfigWithSingleBucket(size, name, description, zone, accessKey, secretKey string) string {
	return fmt.Sprintf(`
		resource "upcloud_object_storage" "my_storage" {
			size  = %s
			name  = "%s"
			description = "%s"
			zone  = "%s"
			access_key = "%s"
			secret_key = "%s"

			bucket {
				name = "test"
			}
		}
`, size, name, description, zone, accessKey, secretKey)
}

func testUpCloudObjectStorageInstanceDefaultsConfig(size, name, zone, accessKey, secretKey string) string {
	return fmt.Sprintf(`
		resource "upcloud_object_storage" "my_storage" {
			size  = %s
			name = "%s"
			zone  = "%s"
			access_key = "%s"
			secret_key = "%s"
		}
`, size, name, zone, accessKey, secretKey)
}

func testUpCloudObjectStorageWithBucketsInstanceConfig(size, name, zone, accessKey, secretKey string, buckets ...string) string {
	bucketClauses := make([]string, 0, len(buckets))
	for _, bucket := range buckets {
		bucketClause := fmt.Sprintf(`
			bucket {
				name = "%s"
			}
`, bucket)

		bucketClauses = append(bucketClauses, bucketClause)
	}

	return fmt.Sprintf(`
		resource "upcloud_object_storage" "my_storage" {
			size  = %s
			name = "%s"
			zone  = "%s"
			access_key = "%s"
			secret_key = "%s"
			%s
		}
`, size, name, zone, accessKey, secretKey, strings.Join(bucketClauses, "\n\n"))
}

func verifyObjectStorageExists(accessKey, secretKey, name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		exists, err := doesObjectStorageExists(state, accessKey, secretKey)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("could not find instance %s", name)
		}
		return nil
	}
}

func verifyObjectStorageDoesNotExist(name string) resource.TestCheckFunc {
	/*
			The reason of not using doesObjectStorageExists to check the s3 bucket availability is
			because of a race condition.
		    the s3 endpoint is still available few seconds after the API delete call,
		    that's why we check against the API and not the resource.
	*/
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "upcloud_storage" {
				continue
			}

			client := testAccProvider.Meta().(*service.Service)
			_, err := client.GetObjectStorageDetails(context.Background(), &request.GetObjectStorageDetailsRequest{
				UUID: rs.Primary.ID,
			})
			if err != nil {
				svcErr, ok := err.(*upcloud.Problem)

				if ok && svcErr.Status == http.StatusNotFound {
					return nil
				}
				return err
			}

			if err == nil {
				return fmt.Errorf("[ERROR] found instance %s : %s that should have been deleted", name, rs.Primary.ID)
			}
		}
		return nil
	}
}

func doesObjectStorageExists(state *terraform.State, accessKey, secretKey string) (bool, error) {
	_, err := getMinioConnection(state, accessKey, secretKey)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getMinioConnection(state *terraform.State, accessKey, secretKey string) (*minio.Client, error) {
	resources, ok := state.Modules[0].Resources["upcloud_object_storage.my_storage"]
	if !ok {
		return nil, fmt.Errorf("could not find resources")
	}

	return objectstorage.GetBucketConnection(resources.Primary.Attributes["url"], accessKey, secretKey)
}

func verifyBucketExists(accessKey, secretKey, bucketName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		found, err := bucketExists(state, accessKey, secretKey, bucketName)
		if err != nil {
			return err
		}

		if !found {
			return fmt.Errorf("could not find bucket %s", bucketName)
		}
		return nil
	}
}

func verifyBucketDoesNotExist(accessKey, secretKey, bucketName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		found, err := bucketExists(state, accessKey, secretKey, bucketName)
		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("found unexpected bucket %s", bucketName)
		}
		return nil
	}
}

func bucketExists(state *terraform.State, accessKey, secretKey, bucketName string) (bool, error) {
	minio, err := getMinioConnection(state, accessKey, secretKey)
	if err != nil {
		return false, err
	}

	return minio.BucketExists(context.Background(), bucketName)
}
