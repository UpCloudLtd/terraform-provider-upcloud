package upcloud

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/minio/minio-go/v7"
)

const expectedDescription = "My object storage"
const expectedZone = "pl-waw1"
const expectedKey = "an access key"
const expectedSecret = "a secret key"

const expectedName1 = "test-name1"
const expectedName2 = "test-name2"
const expectedName3 = "test-name3"

func init() {
	resource.AddTestSweepers("object_storage_cleanup", &resource.Sweeper{
		Name: "object_storage_cleanup",
		F: func(region string) error {
			var nameMap = map[string]interface{}{
				expectedName1: nil,
				expectedName2: nil,
				expectedName3: nil,
			}

			username, ok := os.LookupEnv("UPCLOUD_USERNAME")
			if !ok {
				return fmt.Errorf("UPCLOUD_USERNAME must be set for acceptance tests")
			}

			password, ok := os.LookupEnv("UPCLOUD_PASSWORD")
			if !ok {
				return fmt.Errorf("UPCLOUD_PASSWORD must be set for acceptance tests")
			}

			client := retryablehttp.NewClient()

			service := newUpCloudServiceConnection(username, password, client.HTTPClient)

			objectStorages, err := service.GetObjectStorages()
			if err != nil {
				return err
			}

			for _, objectStorage := range objectStorages.ObjectStorages {
				_, found := nameMap[objectStorage.Name]
				if !found {
					continue
				}

				err = service.DeleteObjectStorage(&request.DeleteObjectStorageRequest{
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

// TestMain is boilerplate needed for -sweep command line parameters to work.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestUpCloudObjectStorage_basic(t *testing.T) {
	var providers []*schema.Provider

	const expectedSize = "250"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      verifyObjectStorageDoesNotExist(expectedKey, expectedSecret, expectedName1),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedSize, expectedName1, expectedDescription, expectedZone, expectedKey, expectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "name"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "description"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", expectedName1),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName1),
				),
			},
		},
	})
}

func TestUpCloudObjectStorage_basic_update(t *testing.T) {
	var providers []*schema.Provider

	const expectedSize = "500"

	const expectedUpdatedSize = "1000"
	const expectedUpdatedDescription = "My Updated data collection"
	const expectedUpdatedKey = "an updated access key"
	const expectedUpdatedSecret = "an updated secret"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      verifyObjectStorageDoesNotExist(expectedUpdatedKey, expectedUpdatedSecret, expectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedSize, expectedName2, expectedDescription, expectedZone, expectedKey, expectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", expectedName2),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName2),
				),
			},
			{
				Config: testUpCloudObjectStorageInstanceConfig(
					expectedUpdatedSize,
					expectedName2,
					expectedUpdatedDescription,
					expectedZone,
					expectedUpdatedKey,
					expectedUpdatedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedUpdatedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedUpdatedDescription),
					verifyObjectStorageExists(expectedUpdatedKey, expectedUpdatedSecret, expectedName2),
				),
			},
		},
	})
}

func TestUpCloudObjectStorage_default_values(t *testing.T) {
	var providers []*schema.Provider

	const expectedSize = "500"
	const expectedUpdatedSize = "1000"
	const expectedUpdatedKey = "an updated access key"
	const expectedUpdatedSecret = "an updated secret"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      verifyObjectStorageDoesNotExist(expectedUpdatedKey, expectedUpdatedSecret, expectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageInstanceDefaultsConfig(
					expectedSize, expectedName3, expectedZone, expectedKey, expectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
				),
			},
			{
				Config: testUpCloudObjectStorageInstanceDefaultsConfig(
					expectedUpdatedSize,
					expectedName3,
					expectedZone,
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
	var providers []*schema.Provider

	const expectedSize = "500"
	const expectedBucketName1 = "bucket1"
	const expectedBucketName2 = "bucket2"
	const expectedBucketName3 = "bucket3"
	const expectedBucketName4 = "bucket4"
	const expectedBucketName5 = "bucket5"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      verifyObjectStorageDoesNotExist(expectedKey, expectedSecret, expectedName2),
		Steps: []resource.TestStep{
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, expectedName2, expectedZone,
					expectedKey, expectedSecret, expectedBucketName1, expectedBucketName2,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", expectedName2),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName2),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName1),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName2),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, expectedName2, expectedZone,
					expectedKey, expectedSecret, expectedBucketName1,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName2),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName1),
					verifyBucketDoesNotExist(expectedKey, expectedSecret, expectedName2, expectedBucketName2),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, expectedName2, expectedZone,
					expectedKey, expectedSecret, expectedBucketName1, expectedBucketName3, expectedBucketName4,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName2),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName1),
					verifyBucketDoesNotExist(expectedKey, expectedSecret, expectedName2, expectedBucketName2),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName3),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName4),
				),
			},
			{
				Config: testUpCloudObjectStorageWithBucketsInstanceConfig(
					expectedSize, expectedName2, expectedZone,
					expectedKey, expectedSecret, expectedBucketName4, expectedBucketName5,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName2),
					verifyBucketDoesNotExist(expectedKey, expectedSecret, expectedName2, expectedBucketName1),
					verifyBucketDoesNotExist(expectedKey, expectedSecret, expectedName2, expectedBucketName2),
					verifyBucketDoesNotExist(expectedKey, expectedSecret, expectedName2, expectedBucketName3),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName4),
					verifyBucketExists(expectedKey, expectedSecret, expectedName2, expectedBucketName5),
				),
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

func verifyObjectStorageDoesNotExist(accessKey, secretKey, name string) resource.TestCheckFunc {
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
			_, err := client.GetObjectStorageDetails(&request.GetObjectStorageDetailsRequest{
				UUID: rs.Primary.ID,
			})

			if err != nil {
				var svcErr *upcloud.Error

				if errors.As(err, &svcErr) && svcErr.ErrorCode == "404" {
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

	return getBucketConnection(resources.Primary.Attributes["url"], accessKey, secretKey)
}

func verifyBucketExists(accessKey, secretKey, name, bucketName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		found, err := bucketExists(state, accessKey, secretKey, name, bucketName)
		if err != nil {
			return err
		}

		if !found {
			return fmt.Errorf("could not find bucket %s", bucketName)
		}
		return nil
	}
}

func verifyBucketDoesNotExist(accessKey, secretKey, name, bucketName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		found, err := bucketExists(state, accessKey, secretKey, name, bucketName)
		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("found unexpected bucket %s", bucketName)
		}
		return nil
	}
}

func bucketExists(state *terraform.State, accessKey, secretKey, name, bucketName string) (bool, error) {
	minio, err := getMinioConnection(state, accessKey, secretKey)
	if err != nil {
		return false, err
	}

	return minio.BucketExists(context.Background(), bucketName)
}
