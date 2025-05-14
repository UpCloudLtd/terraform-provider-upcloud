package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUpcloudManagedObjectStorageRegions(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/data_source_managed_object_storage_regions_s1.tf")

	name := "data.upcloud_managed_object_storage_regions.this"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "regions.#"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "regions.*", map[string]string{
						"name":         "europe-1",
						"primary_zone": "fi-hel2",
					}),
				),
			},
		},
	})
}
