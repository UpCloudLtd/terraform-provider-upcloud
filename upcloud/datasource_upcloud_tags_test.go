package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceUpCloudTags_basic(t *testing.T) {
	resourceName := "data.upcloud_tags.empty"
	tagName := fmt.Sprintf("tag-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudTagsConfigEmpty(tagName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceUpCloudTagsCheck(resourceName),
				),
			},
		},
	})
}

func testAccDataSourceUpCloudTagsCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		_, tagsOk := rs.Primary.Attributes["tags.#"]

		if !tagsOk {
			return fmt.Errorf("tags attribute is missing")
		}

		return nil
	}
}

func testAccDataSourceUpCloudTagsConfigEmpty(tagName string) string {
	return fmt.Sprintf(`
resource "upcloud_tag" "empty" {

  name = "%s"
  description = "A tag for testing"
  servers = []
}

data "upcloud_tags" "empty" {}
`, tagName)
}
