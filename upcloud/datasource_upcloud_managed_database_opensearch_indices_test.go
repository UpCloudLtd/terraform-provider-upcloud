package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceUpcloudManagedDatabaseOpenSearchIndices(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_database/datasource_opensearch_properties_s1.tf")

	var providers []*schema.Provider
	name := "data.upcloud_managed_database_opensearch_indices.datasource_opensearch_properties_s1"
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
					resource.TestCheckResourceAttrSet(name, prop("create_time")),
					resource.TestCheckResourceAttr(name, prop("docs"), "0"),
					resource.TestCheckResourceAttr(name, prop("health"), "green"),
					resource.TestCheckResourceAttr(name, prop("index_name"), ".kibana_1"),
					resource.TestCheckResourceAttr(name, prop("number_of_replicas"), "0"),
					resource.TestCheckResourceAttr(name, prop("number_of_shards"), "1"),
					resource.TestCheckResourceAttr(name, prop("read_only_allow_delete"), "false"),
					resource.TestCheckResourceAttr(name, prop("size"), "208"),
					resource.TestCheckResourceAttr(name, prop("status"), "open"),
				),
			},
		},
	})
}
