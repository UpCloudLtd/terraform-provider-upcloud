package managedobjectstoragetests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedObjectStorageUser(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_user_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_user_s2.tf")

	storage := "upcloud_managed_object_storage.user"
	policy := "upcloud_managed_object_storage_policy.user"
	escapePolicy := "upcloud_managed_object_storage_policy.escape"
	user := "upcloud_managed_object_storage_user.user"
	userAccessKey := "upcloud_managed_object_storage_user_access_key.user"
	userPolicy := "upcloud_managed_object_storage_user_policy.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-iam-objsto"),
					resource.TestCheckResourceAttr(policy, "name", "get-user-policy"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-iam-user"),
					resource.TestCheckResourceAttr(userAccessKey, "username", "tf-acc-test-objstov2-iam-user"),
					resource.TestCheckResourceAttr(userAccessKey, "status", "Active"),
					resource.TestCheckResourceAttr(userPolicy, "username", "tf-acc-test-objstov2-iam-user"),
					resource.TestCheckResourceAttr(userPolicy, "name", "get-user-policy"),
				),
			},
			{
				// Validate that there are no changes planned (to e.g. policy document because of server side formatting changes)
				Config:   testDataS1,
				PlanOnly: true,
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
				ResourceName:      escapePolicy,
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
				Config:            testDataS1,
				ResourceName:      userAccessKey,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"last_used_at",
					"secret_access_key", // only provided on creation, not available on subsequent requests like import
				},
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-iam-objsto"),
					resource.TestCheckResourceAttr(policy, "name", "get-user-policy"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-iam-user"),
					resource.TestCheckResourceAttr(userAccessKey, "username", "tf-acc-test-objstov2-iam-user"),
					resource.TestCheckResourceAttr(userAccessKey, "status", "Inactive"),
				),
			},
		},
	})
}
