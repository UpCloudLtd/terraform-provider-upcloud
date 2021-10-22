package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckServerSimpleBackup(serverName, simpleBackupRule string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*service.Service)
		resourceName := fmt.Sprintf("upcloud_server.%s", serverName)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Server with the name: %s is not set", serverName)
		}

		serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{UUID: rs.Primary.ID})
		if err != nil {
			return err
		}

		if serverDetails.SimpleBackup != simpleBackupRule {
			return fmt.Errorf("Server simple backup rule does not match. Expected: %s, received: %s", simpleBackupRule, serverDetails.SimpleBackup)
		}

		return nil
	}
}

func TestAccUpCloudServerBackup_basic(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				// basic setp
				Config: `
					resource "upcloud_server" "x1" {
						zone = "fi-hel1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}
					
					resource "upcloud_server_backup" "sb1" {
						server = upcloud_server.x1.id
						plan = "weeklies"
						time = "2200"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "plan", "weeklies"),
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "time", "2200"),
					testAccCheckServerSimpleBackup("x1", "2200,weeklies"),
				),
			},
			{
				// test update
				Config: `
					resource "upcloud_server" "x1" {
						zone = "fi-hel1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}
					
					resource "upcloud_server_backup" "sb1" {
						server = upcloud_server.x1.id
						plan = "dailies"
						time = "0000"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "plan", "dailies"),
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "time", "0000"),
					testAccCheckServerSimpleBackup("x1", "0000,dailies"),
				),
			},
			{
				// test deletion
				Config: `
					resource "upcloud_server" "x1" {
						zone = "fi-hel1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"

						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}

						network_interface {
							type = "public"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerSimpleBackup("x1", "no"),
				),
			},
		},
	})
}

func TestAccUpCloudServerBackup_withServerSwitching(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_server" "x1" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}
					
					resource "upcloud_server" "x2" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}

					resource "upcloud_server_backup" "sb1" {
						server = upcloud_server.x1.id
						plan = "weeklies"
						time = "2200"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "plan", "weeklies"),
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "time", "2200"),
					testAccCheckServerSimpleBackup("x1", "2200,weeklies"),
				),
			},
			{
				Config: `
					resource "upcloud_server" "x1" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}
					
					resource "upcloud_server" "x2" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "mainx1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					}

					resource "upcloud_server_backup" "sb1" {
						server = upcloud_server.x2.id
						plan = "weeklies"
						time = "2200"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "plan", "weeklies"),
					resource.TestCheckResourceAttr("upcloud_server_backup.sb1", "time", "2200"),
					testAccCheckServerSimpleBackup("x2", "2200,weeklies"),
					testAccCheckServerSimpleBackup("x1", "no"),
				),
			},
		},
	})
}
