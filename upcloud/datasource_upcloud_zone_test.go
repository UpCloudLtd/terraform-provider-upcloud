package upcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	configWithName = `
data "upcloud_zone" "my_zone" {
  name = "uk-lon1"
}`
	configWithID = `
data "upcloud_zone" "my_zone" {
  id = "uk-lon1"
}`
	configWithNameAndID = `
data "upcloud_zone" "my_zone" {
  id = "uk-lon1"
  name = "de-fra1"
}`
)

func TestAccDataSourceUpCloudZone_basic(t *testing.T) {
	resourceName := "data.upcloud_zone.my_zone"

	expectedZoneName := "uk-lon1"
	expectedDescription := "London #1"
	expectedPublic := "true"

	var steps []resource.TestStep
	for _, config := range []string{
		configWithName,
		configWithID,
		configWithNameAndID,
	} {
		steps = append(steps, resource.TestStep{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(
					resourceName, "name", expectedZoneName),
				resource.TestCheckResourceAttr(
					resourceName, "description", expectedDescription),
				resource.TestCheckResourceAttr(
					resourceName, "public", expectedPublic),
			),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps:                    steps,
	})
}
