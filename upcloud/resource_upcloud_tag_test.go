package upcloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudTag_basic(t *testing.T) {
	var providers []*schema.Provider

	expectedNames := []string{"dev", "test"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_dev", "name"),
					resource.TestCheckResourceAttr("upcloud_tag.my_tag_dev", "name", expectedNames[0]),
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_test", "name"),
					resource.TestCheckResourceAttr("upcloud_tag.my_tag_test", "name", expectedNames[1]),
				),
			},
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_dev", "name"),
					resource.TestCheckResourceAttr("upcloud_tag.my_tag_dev", "name", expectedNames[0]),
				),
			},
		},
	})
}

func testUpcloudTagInstanceConfig(names []string) string {

	config := strings.Builder{}

	for _, name := range names {
		config.WriteString(fmt.Sprintf(`
		resource "upcloud_tag" "my_tag_%s" {
  			name = "%s"
  			description = "Represents the %s environment"
  			servers = []
		}`, name, name, name))
	}

	return config.String()
}
