package upcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUpCloudStorage(t *testing.T) {
	templateResourceName := "data.upcloud_storage.ubuntu_template"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: dataSourceUpCloudStorageTestTemplateConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(templateResourceName, "id", "01000000-0000-4000-8000-000030200200"),
					resource.TestCheckResourceAttr(templateResourceName, "type", "template"),
					resource.TestCheckResourceAttr(templateResourceName, "name", "Ubuntu Server 20.04 LTS (Focal Fossa)"),
					resource.TestCheckResourceAttr(templateResourceName, "access_type", "public"),
					resource.TestCheckResourceAttr(templateResourceName, "size", "4"),
					resource.TestCheckResourceAttr(templateResourceName, "state", "online"),
					resource.TestCheckResourceAttr("data.upcloud_storage.regex", "id", "01000000-0000-4000-8000-000020060100"),
				),
			},
			{
				Config: dataSourceUpCloudStorageTestConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage.hel1", "title", "hel1"),
					resource.TestCheckResourceAttr("upcloud_storage.hel2", "title", "hel2"),
					resource.TestCheckResourceAttr("data.upcloud_storage.most_recent", "title", "hel2"),
					resource.TestCheckResourceAttr("data.upcloud_storage.name", "title", "hel1"),
					resource.TestCheckResourceAttr("data.upcloud_storage.zone", "title", "hel1"),
					resource.TestCheckResourceAttr("data.upcloud_storage.most_recent_private", "title", "hel2"),
				),
			},
		},
	})
}

func dataSourceUpCloudStorageTestTemplateConfig() string {
	return `
	data "upcloud_storage" "ubuntu_template" {
		type = "template"
		name = "Ubuntu Server 20.04 LTS (Focal Fossa)"
	}
	data "upcloud_storage" "regex" {
		type = "template"
		access_type = "public"
		name_regex = "^Debian GNU/Linux [0-1]{2} \\(Bullseye\\)$"
	}
	`
}

func dataSourceUpCloudStorageTestConfig() string {
	return `
	resource "upcloud_storage" "hel1" {
		size  = 10
		tier  = "maxiops"
		title = "hel1"
		zone  = "fi-hel1"
	}
	resource "upcloud_storage" "hel2" {
		size  = 10
		tier  = "maxiops"
		title = "hel2"
		zone  = "fi-hel2"
		# depend on hel1 to put more time between 'created' time to test 'most_recent' filter
		depends_on = [upcloud_storage.hel1]
	}
	# most recent storage
	data "upcloud_storage" "most_recent" {
		type = "normal"
		name_regex = "^hel"
		most_recent = true
		depends_on = [
			upcloud_storage.hel1,
			upcloud_storage.hel2
		]
	}
	# use exact name
	data "upcloud_storage" "name" {
		type = "normal"
		name = "hel1"
		depends_on = [
			upcloud_storage.hel1,
			upcloud_storage.hel2
		]
	}
	# most recent storage per zone
	data "upcloud_storage" "zone" {
		type = "normal"
		name_regex = "^hel"
		most_recent = true
		zone = "fi-hel1"
		depends_on = [
			upcloud_storage.hel1,
			upcloud_storage.hel2
		]
	}
	# most recent prive disk
	data "upcloud_storage" "most_recent_private" {
		type = "normal"
		name_regex = ".*"
		access_type = "private"
		most_recent = true
		depends_on = [
			upcloud_storage.hel1,
			upcloud_storage.hel2
		]
	}
	`
}
