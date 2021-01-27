package upcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net/url"
	"testing"
	"time"
)

const expectedDescription = "My object storage"
const expectedZone = "fi-hel2"
const expectedKey = "an access key"
const expectedSecret = "a secret key"

func TestUpcloudObjectStorage_basic(t *testing.T) {
	var providers []*schema.Provider

	const expectedSize = "250"
	const expectedName = "name1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy: verifyObjectStorageDoesNotExist(expectedKey, expectedSecret, expectedName),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudObjectStorageInstanceConfig(
					expectedSize, expectedName, expectedDescription, expectedZone, expectedKey, expectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "name"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "description"),
					resource.TestCheckResourceAttrSet("upcloud_object_storage.my_storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", expectedName),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName),
				),
			},
		},
	})
}

func TestUpcloudObjectStorage_basic_update(t *testing.T) {
	var providers []*schema.Provider

	const expectedSize = "500"
	const expectedName = "name2"

	const expectedUpdatedSize = "1000"
	const expectedUpdatedDescription = "My Updated data collection"
	const expectedUpdatedKey = "an updated access key"
	const expectedUpdatedSecret = "an updated secret"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy: verifyObjectStorageDoesNotExist(expectedUpdatedKey, expectedUpdatedSecret, expectedName),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudObjectStorageInstanceConfig(
					expectedSize, expectedName, expectedDescription, expectedZone, expectedKey, expectedSecret,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "size", expectedSize),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "name", expectedName),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "description", expectedDescription),
					resource.TestCheckResourceAttr(
						"upcloud_object_storage.my_storage", "zone", expectedZone),
					verifyObjectStorageExists(expectedKey, expectedSecret, expectedName),
				),
			},
			{
				Config: testUpcloudObjectStorageInstanceConfig(
					expectedUpdatedSize,
					expectedName,
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
					verifyObjectStorageExists(expectedUpdatedKey, expectedUpdatedSecret, expectedName),
				),
			},
		},
	})
}

func testUpcloudObjectStorageInstanceConfig(size, name, description, zone, access_key, secret_key string) string {
	return fmt.Sprintf(`
		resource "upcloud_object_storage" "my_storage" {
			size  = %s
			name  = "%s"
			description = "%s"
			zone  = "%s"
			access_key = "%s"
			secret_key = "%s"
		}
`, size, name, description, zone, access_key, secret_key)
}

func verifyObjectStorageExists(accessKey, secretKey, name string) resource.TestCheckFunc {
	return func (state * terraform.State) error {
		exists, err := doesObjectStorageExists(state, accessKey, secretKey, name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("could not find bucket %s", name)
		}
		return nil
	}
}

func verifyObjectStorageDoesNotExist(accessKey, secretKey, name string) resource.TestCheckFunc {
	return func (state * terraform.State) error {
		time.Sleep(time.Second * 3)
		exists, err := doesObjectStorageExists(state, accessKey, secretKey, name)
		if err != nil {
			if err.Error() == "could not find resources" {
				return nil
			}
			return err
		}
		if exists {
			return fmt.Errorf("found bucket %s that should have been deleted", name)
		}
		return nil
	}
}

func doesObjectStorageExists(state * terraform.State, accessKey, secretKey, name string) (bool, error) {
	resources, ok := state.Modules[0].Resources["upcloud_object_storage.my_storage"]
	if !ok {
		return false, fmt.Errorf("could not find resources")
	}

	url, err := url.Parse(resources.Primary.Attributes["url"])
	if err != nil {
		return false, err
	}

	_, err = minio.New(url.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
