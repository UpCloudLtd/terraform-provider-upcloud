package upcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestUpcloudStorage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUpcloudStorageInstanceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_storage.my-storage", "size"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my-storage", "tier"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my-storage", "title"),
					resource.TestCheckResourceAttrSet("upcloud_storage.my-storage", "zone"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my-storage", "size", "10"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my-storage", "tier", "maxiops"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my-storage", "title", "My data collection"),
					resource.TestCheckResourceAttr(
						"upcloud_storage.my-storage", "zone", "fi-hel1"),
				),
			},
		},
	})
}

func testUpcloudStorageInstanceConfig() string {
	return fmt.Sprintf(`
		resource "upcloud_storage" "my-storage" {
			size  = 10
			tier  = "maxiops"
			title = "My data collection"
			zone  = "fi-hel1"
		}
`)
}
