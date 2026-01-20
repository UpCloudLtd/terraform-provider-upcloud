package tagtests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	upc "github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUpcloudTag_basic(t *testing.T) {
	tag1 := acctest.RandString(10)
	tag2 := acctest.RandString(10)
	expectedNames := []string{tag1, tag2}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upc.TestAccProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_1", "name"),
					resource.TestCheckResourceAttr(
						"upcloud_tag.my_tag_1",
						"name",
						expectedNames[0],
					),
					resource.TestCheckResourceAttrSet("upcloud_tag.my_tag_2", "name"),
					resource.TestCheckResourceAttr(
						"upcloud_tag.my_tag_2",
						"name",
						expectedNames[1],
					),
				),
			},
		},
	})
}

func TestAccUpCloudTag_import(t *testing.T) {
	var tags upcloud.Tags

	tag1 := acctest.RandString(10)
	expectedNames := []string{tag1}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upc.TestAccProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudTagInstanceConfig(expectedNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagsExists("upcloud_tag.my_tag_1", &tags),
				),
			},
			{
				ResourceName:      "upcloud_tag.my_tag_1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTagsExists(resourceName string, tags *upcloud.Tags) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Look for the full resource name and error if not found
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// The provider has not set the ID for the resource
		if rs.Primary.ID == "" {
			return fmt.Errorf("No tag ID is set")
		}

		// Use the API SDK to locate the remote resource.
		client := upc.TestAccProvider.Meta().(*service.Service)
		latest, err := client.GetTags(context.Background())
		if err != nil {
			return err
		}

		// Update the reference the remote located tags
		*tags = *latest

		return nil
	}
}

func testAccCheckTagDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "upcloud_tag" {
			continue
		}

		client := upc.TestAccProvider.Meta().(*service.Service)
		tags, err := client.GetTags(context.Background())
		if err != nil {
			return fmt.Errorf(
				"[WARN] Error listing tags when deleting upcloud tag (%s): %s",
				rs.Primary.ID,
				err,
			)
		}

		for _, tag := range tags.Tags {
			if tag.Name == rs.Primary.ID {
				return fmt.Errorf(
					"[WARN] Tried deleting tag (%s), but was still found",
					rs.Primary.ID,
				)
			}
		}
	}
	return nil
}

func testUpcloudTagInstanceConfig(names []string) string {
	config := strings.Builder{}

	for idx, name := range names {
		config.WriteString(fmt.Sprintf(`
		resource "upcloud_tag" "my_tag_%s" {
  			name = "%s"
  			description = "Represents the %s environment"
		}`, fmt.Sprint(idx+1), name, name))
	}

	return config.String()
}
