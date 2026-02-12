package databasetests

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedDatabaseOpenSearchProperties(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_database/opensearch_properties_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_database/opensearch_properties_s2.tf")

	name := "upcloud_managed_database_opensearch.opensearch_properties"
	prop := func(name string) string {
		return fmt.Sprintf("properties.0.%s", name)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "1x2xCPU-4GB-80GB-1D"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(name, "access_control", "false"),
					resource.TestCheckResourceAttr(name, "extended_access_control", "false"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "false"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "false"),
					resource.TestCheckResourceAttr(name, prop("version"), "2.19"),
				),
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "plan", "1x2xCPU-4GB-80GB-1D"),
					resource.TestCheckResourceAttr(name, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(name, "access_control", "true"),
					resource.TestCheckResourceAttr(name, "extended_access_control", "true"),
					resource.TestCheckResourceAttr(name, prop("public_access"), "true"),
					resource.TestCheckResourceAttr(name, prop("automatic_utility_network_ip_filter"), "true"),
					resource.TestCheckResourceAttr(name, prop("version"), "2.19"),
				),
			},
		},
	})
}
