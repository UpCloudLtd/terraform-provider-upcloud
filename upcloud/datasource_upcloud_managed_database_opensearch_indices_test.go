package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceUpcloudManagedDatabaseOpenSearchIndices(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/data_source_opensearch_indices_s1.tf")

	var providers []*schema.Provider
	name := "data.upcloud_managed_database_opensearch_indices.opensearch_indices"
	prop := func(name string) string {
		return fmt.Sprintf("indices.0.%s", name)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					// Check these from the first index
					resource.TestCheckResourceAttrSet(name, prop("create_time")),
					resource.TestCheckResourceAttrSet(name, prop("docs")),
					resource.TestCheckResourceAttrSet(name, prop("size")),
					// Check rest of the fields from ".opendistro_security" index
					resource.TestCheckTypeSetElemNestedAttrs(name, "indices.*", map[string]string{
						"index_name":             ".opendistro_security",
						"health":                 "green",
						"number_of_replicas":     "0",
						"number_of_shards":       "1",
						"read_only_allow_delete": "false",
						"status":                 "open",
					}),
				),
			},
		},
	})
}
