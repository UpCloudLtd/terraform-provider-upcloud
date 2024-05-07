package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUpcloudManagedObjectStoragePolicies(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/data_source_managed_object_storage_policies_s1.tf")

	name := "data.upcloud_managed_object_storage_policies.this"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "policies.#"),
				),
			},
		},
	})
}
