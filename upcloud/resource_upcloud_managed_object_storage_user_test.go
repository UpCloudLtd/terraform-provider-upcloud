package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudManagedObjectStorageUser(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_user_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_user_s2.tf")

	var providers []*schema.Provider
	storage := "upcloud_managed_object_storage.user"
	policy := "upcloud_managed_object_storage_policy.user"
	user := "upcloud_managed_object_storage_user.user"
	userPolicy := "upcloud_managed_object_storage_user_policy.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(policy, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userPolicy, "username", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(userPolicy, "name", "tf-acc-test-objstov2-user"),
				),
			},
			{
				Config:            testDataS2,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(storage, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(policy, "name", "tf-acc-test-objstov2-user"),
					resource.TestCheckResourceAttr(user, "username", "tf-acc-test-objstov2-user"),
				),
			},
		},
	})
}
