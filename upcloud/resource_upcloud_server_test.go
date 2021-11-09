package upcloud

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestUpcloudServer_minimal(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_server" "min" {
            hostname = "min-server" 
						zone     = "fi-hel2"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.min", "zone", "fi-hel2"),
					resource.TestCheckResourceAttr("upcloud_server.min", "hostname", "min-server"),
					resource.TestCheckNoResourceAttr("upcloud_server.min", "tags"),
				),
			},
			{
				Config: `
					resource "upcloud_server" "min" {
            hostname = "min-server" 
						zone     = "fi-hel2"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				ExpectNonEmptyPlan: false, //ensure nothing changed
			},
		},
	})
}

func TestUpcloudServer_basic(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"
						title    = "Debian"
						tags = [
						"foo",
						"bar"
						]

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_server.my-server", "zone"),
					resource.TestCheckResourceAttrSet("upcloud_server.my-server", "hostname"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "hostname", "debian.example.com"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.0", "foo",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.1", "bar",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "title", "Debian",
					),
				),
			},
		},
	})
}

func TestUpcloudServer_changePlan(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfigWithSmallServerPlan,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "1xCPU-2GB"),
				),
			},
			{
				Config: testAccPlanConfigUpdateServerPlan,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
				),
			},
		},
	})
}

func TestUpcloudServer_simpleBackup(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				// basic setup
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}

						simple_backup {
							time = "0300"
							plan = "dailies"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.time", "0300"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.plan", "dailies"),
				),
			},
			{
				// change simple backup config
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}

						simple_backup {
							time = "2200"
							plan = "weeklies"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.time", "2200"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.plan", "weeklies"),
				),
			},
			{
				// replace simple backup with backup rule on the template
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
								backup_rule {
									time = "0010"
									interval = "mon"
									retention = 2
								}
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.time", "0010"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.retention", "2"),
					resource.TestCheckNoResourceAttr("upcloud_server.my-server", "simple_backup"),
				),
			},
			{
				// adjust backup rule on the template
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
								backup_rule {
									time = "0010"
									interval = "tue"
									retention = 3
								}
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.time", "0010"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.interval", "tue"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.retention", "3"),
				),
			},
			{
				// replace template backup rule back with simple backup
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						simple_backup {
							time = "2300"
							plan = "dailies"
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.time", "2300"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.plan", "dailies"),
					resource.TestCheckNoResourceAttr("upcloud_server.my-server", "template.0.backup_rule"),
				),
			},
		},
	})
}

func TestUpcloudServer_simpleBackupWithStorage(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				// basic setup
				Config: `
					resource "upcloud_storage" "addon" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
						
						backup_rule {
							time = "0100"
							interval = "mon"
							retention = 2
						}
					}
					
					resource "upcloud_server" "my-server" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "main1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					
						storage_devices {
							storage = upcloud_storage.addon.id
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.time", "0100"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.retention", "2"),
				),
			},
			{
				// replace additional storages backup rule with simple backup
				Config: `
					resource "upcloud_storage" "addon" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "my-server" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "main1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}

						simple_backup {
							time = "2200"
							plan = "dailies"
						}
					
						storage_devices {
							storage = upcloud_storage.addon.id
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.time", "2200"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.plan", "dailies"),
					resource.TestCheckNoResourceAttr("upcloud_storage.addon", "backup_rule"),
				),
			},
			{
				// Update simple backup while storage is attached
				Config: `
					resource "upcloud_storage" "addon" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "my-server" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "main1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}

						simple_backup {
							time = "2300"
							plan = "weeklies"
						}
					
						storage_devices {
							storage = upcloud_storage.addon.id
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.time", "2300"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "simple_backup.0.plan", "weeklies"),
					resource.TestCheckNoResourceAttr("upcloud_storage.addon", "backup_rule"),
				),
			},

			{
				// Delete simple backups while storage is attached
				Config: `
					resource "upcloud_storage" "addon" {
						title = "addon"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "my-server" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "main1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10
						}
					
						network_interface {
							type = "public"
						}
					
						storage_devices {
							storage = upcloud_storage.addon.id
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("upcloud_server.my-server", "simple_backup"),
					resource.TestCheckNoResourceAttr("upcloud_storage.addon", "backup_rule"),
				),
			},

			{
				// Add backup rule to additional storage and to the template
				Config: `
					resource "upcloud_storage" "addon" {
						title = "addon"
						size = 10
						zone = "pl-waw1"

						backup_rule {
							time = "0100"
							interval = "mon"
							retention = 2
						}
					}
					
					resource "upcloud_server" "my-server" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "main1"
						
						template {
							storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
							size = 10

							backup_rule {
								time = "2200"
								interval = "daily"
								retention = 4
							}
						}
					
						network_interface {
							type = "public"
						}
					
						storage_devices {
							storage = upcloud_storage.addon.id
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.time", "2200"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.interval", "daily"),
					resource.TestCheckResourceAttr("upcloud_server.my-server", "template.0.backup_rule.0.retention", "4"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.time", "0100"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.retention", "2"),
				),
			},
		},
	})
}

func TestUpcloudServerUpdateTags(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				// setup server with tags
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"
						tags = [
						"foo",
						"bar"
						]

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.0", "foo",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.1", "bar",
					),
				),
			},
			{
				// tags update
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"
						tags = [
						"newfoo",
						"newbar"
						]

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.0", "newfoo",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.1", "newbar",
					),
				),
			},
			{
				// tag removal
				Config: `
					resource "upcloud_server" "my-server" {
						zone     = "fi-hel1"
						hostname = "debian.example.com"
						tags = [
						"newfoo",
						]

						template {
								storage = "01000000-0000-4000-8000-000020050100"
								size = 10
						}

						network_interface {
							type = "utility"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.0", "newfoo",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "tags.#", "1",
					),
				),
			},
		},
	})
}

func TestUpcloudServer_networkInterface(t *testing.T) {
	var providers []*schema.Provider

	var serverID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					testAccGetServerID("upcloud_server.my-server", &serverID),
				),
			},
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
					networkInterface{
						niType:  "private",
						network: true,
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.ip_address"),
					testAccCheckServerIDNotEqual("upcloud_server.my-server", serverID),
					testAccCheckNetwork("upcloud_server.my-server", 1, "upcloud_network.test_network_1"),
					testAccGetServerID("upcloud_server.my-server", &serverID),
				),
			},
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
					networkInterface{
						niType:     "private",
						network:    true,
						newNetwork: true,
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.my-server",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.my-server",
						"network_interface.1.ip_address"),
					testAccCheckServerIDNotEqual("upcloud_server.my-server", serverID),
					testAccCheckNetwork("upcloud_server.my-server", 1, "upcloud_network.test_network_11"),
				),
			},
		},
	})
}

func testAccGetServerID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*id = s.RootModule().Resources[resourceName].Primary.ID

		return nil
	}
}

func testAccCheckServerIDNotEqual(resourceName string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		newID := s.RootModule().Resources[resourceName].Primary.ID
		if newID == id {
			return fmt.Errorf("new server ID unexpectedly equals old ID: %s == %s", newID, id)
		}

		return nil
	}
}

func testAccCheckNetwork(resourceName string, niIdx int, networkResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		server := s.RootModule().Resources[resourceName]
		network := s.RootModule().Resources[networkResourceName]
		if network == nil {
			return fmt.Errorf("network resource %s not found", networkResourceName)
		}

		serverNetworkID := server.Primary.Attributes[fmt.Sprintf("network_interface.%d.network", niIdx)]
		networkID := network.Primary.ID

		if serverNetworkID != networkID {
			return fmt.Errorf("server network ID and network ID do not match: %s != %s", serverNetworkID, networkID)
		}

		cidrRange := network.Primary.Attributes["ip_network.0.address"]
		serverIPStr := server.Primary.Attributes[fmt.Sprintf("network_interface.%d.ip_address", niIdx)]

		_, ipNet, err := net.ParseCIDR(cidrRange)
		if err != nil {
			return err
		}

		serverIP := net.ParseIP(serverIPStr)
		if !ipNet.Contains(serverIP) {
			return fmt.Errorf("server IP address is not in networks IP range: %s not in %s", serverIPStr, cidrRange)
		}

		return nil
	}
}

const testAccServerConfigWithSmallServerPlan = `
resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "1xCPU-2GB"

			template {
					storage = "01000000-0000-4000-8000-000020050100"
					size = 10
			}

			network_interface {
				type = "utility"
			}
		}
`

const testAccPlanConfigUpdateServerPlan = `
resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "2xCPU-4GB"

			template {
					storage = "01000000-0000-4000-8000-000020050100"
					size = 10
			}

			network_interface {
				type = "utility"
			}
		}
`

type networkInterface struct {
	niType     string
	network    bool
	newNetwork bool
}

func testAccServerNetworkInterfaceConfig(nis ...networkInterface) string {
	var builder strings.Builder

	builder.WriteString(`
		resource "upcloud_server" "my-server" {
			zone     = "fi-hel1"
			hostname = "debian.example.com"
			plan     = "2xCPU-4GB"

			template {
					storage = "01000000-0000-4000-8000-000020050100"
					size = 10
			}
	`)

	for i, ni := range nis {
		builder.WriteString(fmt.Sprintf(`
				network_interface {
					type = "%s"
		`, ni.niType))

		if ni.network && !ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
						network = upcloud_network.test_network_%d.id
			`, i))
		} else if ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
						network = upcloud_network.test_network_%d.id
			`, 10+i))
		}
		builder.WriteString(`
				}
		`)
	}

	builder.WriteString(`
		}
	`)

	for i, ni := range nis {
		netID := rand.Intn(255)
		if ni.network {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "test_network_%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, i, i, netID, netID))
		}

		netID = rand.Intn(255)
		if ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "test_network_%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, 10+i, 10+i, netID, netID))
		}
	}

	return builder.String()
}

func TestCloudServerDefaultTitle(t *testing.T) {
	want := "terraformterraformterraformterraformte... (managed by terraform)"
	got := cloudServerDefaultTitleFromHostname("terraformterraformterraformterraformterraformterraformterraform")
	if want != got {
		t.Errorf("cloudServerDefaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}
}
