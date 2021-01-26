package upcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccDataSourceUpCloudTags_basic(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "data.upcloud_tags.empty"
	tagName := fmt.Sprintf("tag-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUpCloudTagsConfig_empty(tagName),
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
			return fmt.Errorf("tags attribute is missing.")
		}

		return nil
	}
}

func testAccDataSourceUpCloudTagsConfig_empty(tagName string) string {
	return fmt.Sprintf(`
resource "upcloud_tag" "empty" {

  name = "%s"
  description = "A tag for testing"
  servers = []
}

data "upcloud_tags" "empty" {}
`, tagName)
}
