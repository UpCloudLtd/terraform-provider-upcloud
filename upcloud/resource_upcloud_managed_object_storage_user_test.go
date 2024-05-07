package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUpcloudManagedObjectStorageUser(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_user_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_user_s2.tf")

	storage := "upcloud_managed_object_storage.user"
	policy := "upcloud_managed_object_storage_policy.user"
	user := "upcloud_managed_object_storage_user.user"
	userAccessKey := "upcloud_managed_object_storage_user_access_key.user"
	userPolicy := "upcloud_managed_object_storage_user_policy.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(policy, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userAccessKey, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userAccessKey, "status", "Active"),
					resource.TestCheckResourceAttr(userPolicy, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userPolicy, "name", "tf-acc-test-objstov2-user"),
				),
			},
			{
				Config:            testDataS1,
				ResourceName:      user,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:            testDataS1,
				ResourceName:      userPolicy,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:            testDataS1,
				ResourceName:      policy,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:                  testDataS1,
				ResourceName:            userAccessKey,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_access_key"}, // only provided on creation, not available on subsequent requests like import
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(policy, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userAccessKey, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userAccessKey, "status", "Inactive"),
				),
			},
		},
	})
}
