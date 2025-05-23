package upcloud

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/crypto/ssh"
)

func configCustomPlan(cpu, mem int) string {
	return fmt.Sprintf(`
		resource "upcloud_server" "custom" {
			hostname = "tf-acc-test-server-custom-plan" 
			zone     = "fi-hel2"
			cpu      = "%d"
			mem		 = "%d"
			metadata = true

			template {
					storage = "%s"
					size = 10
			}

			network_interface {
				type = "utility"
			}
		}`, cpu, mem, debianTemplateUUID)
}

func TestUpcloudServer_customPlan(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configCustomPlan(1, 1024),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.custom", "plan", "custom"),
					resource.TestCheckResourceAttr("upcloud_server.custom", "cpu", "1"),
					resource.TestCheckResourceAttr("upcloud_server.custom", "mem", "1024"),
				),
			},
			{
				Config: configCustomPlan(2, 2048),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.custom", "plan", "custom"),
					resource.TestCheckResourceAttr("upcloud_server.custom", "cpu", "2"),
					resource.TestCheckResourceAttr("upcloud_server.custom", "mem", "2048"),
				),
			},
		},
	})
}

func TestUpcloudServer_minimal(t *testing.T) {
	config := fmt.Sprintf(`
		resource "upcloud_server" "this" {
			hostname = "tf-acc-test-server-minimal" 
			zone     = "fi-hel2"
			metadata = true
			plan = "1xCPU-1GB" 

			template {
				storage = "%s"
				size = 10
			}

			network_interface {
				type = "utility"
			}
		}`, debianTemplateUUID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "zone", "fi-hel2"),
					resource.TestCheckResourceAttr("upcloud_server.this", "hostname", "tf-acc-test-server-minimal"),
					resource.TestCheckResourceAttr("upcloud_server.this", "tags.#", "0"),
				),
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: false, // ensure nothing changed
			},
		},
	})
}

func TestUpcloudServer_basic(t *testing.T) {
	config := fmt.Sprintf(`
		resource "upcloud_server" "this" {
			zone     = "fi-hel1"
			hostname = "tf-acc-test-server-basic"
			title    = "tf-acc-test-server-basic"
			metadata = true
			plan = "1xCPU-1GB" 

			labels   = {
				env         = "dev",
				production  = "false"
			}

			tags = [
				"foo",
				"bar"
			]

			template {
				encrypt = true
				storage = "%s"
				size = 10
				filesystem_autoresize = true
				delete_autoresize_backup = true
			}

			network_interface {
				type = "utility"
			}
		}`, debianTemplateUUID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.encrypt", "true"),
					resource.TestCheckResourceAttrSet("upcloud_server.this", "zone"),
					resource.TestCheckResourceAttrSet("upcloud_server.this", "hostname"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "zone", "fi-hel1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "hostname", "tf-acc-test-server-basic"),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "foo",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "labels.env", "dev",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "labels.production", "false",
					),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "bar",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "title", "tf-acc-test-server-basic",
					),
				),
			},
		},
	})
}

func configSimple(hostname, plan, zone string) string {
	return fmt.Sprintf(`
	resource "upcloud_server" "this" {
		hostname = "%s"
		plan     = "%s"
		zone     = "%s"
		metadata = true

		template {
			storage = "%s"
			size = 10
		}

		network_interface {
			type = "utility"
		}
	}`, hostname, plan, zone, debianTemplateUUID)
}

func TestUpcloudServer_changePlan(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configSimple("tf-acc-test-server-change-plan", "1xCPU-2GB", "fi-hel1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "plan", "1xCPU-2GB"),
				),
			},
			{
				Config: configSimple("tf-acc-test-server-change-plan", "2xCPU-4GB", "fi-hel1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "plan", "2xCPU-4GB"),
				),
			},
		},
	})
}

func TestUpcloudServer_developerPlan(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configSimple("tf-acc-test-server-dev-plan", "DEV-1xCPU-1GB", "fi-hel1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "plan", "DEV-1xCPU-1GB"),
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.tier", "standard"),
				),
			},
			{
				Config: configSimple("tf-acc-test-server-dev-plan", "1xCPU-1GB", "fi-hel1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "plan", "1xCPU-1GB"),
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.tier", "standard"),
				),
			},
		},
	})
}

func configSimpleBackup(time, plan string) string {
	return fmt.Sprintf(`
		resource "upcloud_server" "this" {
			hostname = "tf-acc-test-server-simple-backup"
			zone     = "fi-hel1"
			metadata = true
			plan = "1xCPU-1GB"

			template {
				storage = "%s"
				size = 10
			}

			network_interface {
				type = "utility"
			}

			simple_backup {
				time = "%s"
				plan = "%s"
			}
		}`, debianTemplateUUID, time, plan)
}

func configBackupRule(time, interval string, retention int) string {
	return fmt.Sprintf(`
		resource "upcloud_server" "this" {
			zone     = "fi-hel1"
			hostname = "tf-acc-test-server-simple-backup"
			metadata = true

			template {
					storage = "%s"
					size = 10
					backup_rule {
						time = "%s"
						interval = "%s"
						retention = %d
					}
			}

			network_interface {
				type = "utility"
			}
		}`, debianTemplateUUID, time, interval, retention)
}

func TestUpcloudServer_simpleBackup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// basic setup
				Config: configSimpleBackup("0300", "dailies"),
				Check: resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "simple_backup.*", map[string]string{
					"time": "0300",
					"plan": "dailies",
				}),
			},
			{
				// change simple backup config
				Config: configSimpleBackup("2200", "weeklies"),
				Check: resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "simple_backup.*", map[string]string{
					"time": "2200",
					"plan": "weeklies",
				}),
			},
			{
				// replace simple backup with backup rule on the template
				Config: configBackupRule("0010", "mon", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "simple_backup.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "template.0.backup_rule.*", map[string]string{
						"time":      "0010",
						"interval":  "mon",
						"retention": "2",
					}),
				),
			},
			{
				// adjust backup rule on the template
				Config: configBackupRule("0010", "tue", 3),
				Check: resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "template.0.backup_rule.*", map[string]string{
					"time":      "0010",
					"interval":  "tue",
					"retention": "3",
				}),
			},
			{
				// replace template backup rule back with simple backup
				Config: configSimpleBackup("2300", "daily"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.backup_rule.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "simple_backup.*", map[string]string{
						"time": "2300",
						"plan": "daily",
					}),
				),
			},
		},
	})
}

func TestUpcloudServer_simpleBackupWithStorage(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// basic setup
				Config: `
					resource "upcloud_storage" "addon" {
						title = "tf-acc-test-server-storage-simple-backup-extra-disk"
						size = 10
						zone = "pl-waw1"
						
						backup_rule {
							time = "0100"
							interval = "mon"
							retention = 2
						}
					}
					
					resource "upcloud_server" "this" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "tf-acc-test-server-storage-simple-backup"
						
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
				Check: resource.TestCheckTypeSetElemNestedAttrs("upcloud_storage.addon", "backup_rule.*", map[string]string{
					"time":      "0100",
					"interval":  "mon",
					"retention": "2",
				}),
			},
			{
				// replace additional storages backup rule with simple backup
				Config: `
					resource "upcloud_storage" "addon" {
						title = "tf-acc-test-server-storage-simple-backup-extra-disk"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "this" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "tf-acc-test-server-storage-simple-backup"
						
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
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "simple_backup.*", map[string]string{
						"time": "2200",
						"plan": "dailies",
					}),
				),
			},
			{
				// Update simple backup while storage is attached
				Config: `
					resource "upcloud_storage" "addon" {
						title = "tf-acc-test-server-storage-simple-backup-extra-disk"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "this" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "tf-acc-test-server-storage-simple-backup"
						
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
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("upcloud_server.this", "simple_backup.*", map[string]string{
						"time": "2300",
						"plan": "weeklies",
					}),
				),
			},

			{
				// Delete simple backups while storage is attached
				Config: `
					resource "upcloud_storage" "addon" {
						title = "tf-acc-test-server-storage-simple-backup-extra-disk"
						size = 10
						zone = "pl-waw1"
					}
					
					resource "upcloud_server" "this" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "tf-acc-test-server-storage-simple-backup"
						
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
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.#", "0"),
					resource.TestCheckResourceAttr("upcloud_server.this", "simple_backup.#", "0"),
				),
			},

			{
				// Add backup rule to additional storage and to the template
				Config: `
					resource "upcloud_storage" "addon" {
						title = "tf-acc-test-server-storage-simple-backup-extra-disk"
						size = 10
						zone = "pl-waw1"

						backup_rule {
							time = "0100"
							interval = "mon"
							retention = 2
						}
					}
					
					resource "upcloud_server" "this" {
						zone = "pl-waw1"
						plan = "1xCPU-1GB"
						hostname = "tf-acc-test-server-storage-simple-backup"
						
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
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.backup_rule.0.time", "2200"),
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.backup_rule.0.interval", "daily"),
					resource.TestCheckResourceAttr("upcloud_server.this", "template.0.backup_rule.0.retention", "4"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.time", "0100"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.interval", "mon"),
					resource.TestCheckResourceAttr("upcloud_storage.addon", "backup_rule.0.retention", "2"),
				),
			},
		},
	})
}

func configTags(tags ...string) string {
	tagsStr := ""
	if len(tags) > 0 {
		tagsStr = fmt.Sprintf(`"%s"`, strings.Join(tags, `", "`))
	}

	return fmt.Sprintf(`
		resource "upcloud_server" "this" {
			hostname = "tf-acc-test-server-tags"
			plan     = "1xCPU-1GB"
			zone     = "fi-hel1"
			metadata = true
			tags     = [%s]

			template {
				storage = "%s"
				size = 10
			}

			network_interface {
				type = "utility"
			}
		}`, tagsStr, debianTemplateUUID)
}

func TestUpcloudServer_updateTags(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Setup server with tags
				Config: configTags("acceptance-test", "foo", "bar"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "acceptance-test",
					),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "foo",
					),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "bar",
					),
				),
			},
			{
				// Update some of the tags
				Config: configTags("acceptance-test", "newfoo", "newbar"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "acceptance-test",
					),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "newfoo",
					),
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "newbar",
					),
				),
			},
			{
				// Remove some of the tags
				Config: configTags("acceptance-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"upcloud_server.this", "tags.*", "acceptance-test",
					),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "tags.#", "1",
					),
				),
			},
			{
				// Remove all of the tags
				Config: configTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "tags.#", "0",
					),
				),
			},
			{
				// Remove tags attribute
				Config: configSimple("tf-acc-test-server-tags", "1xCPU-1GB", "fi-hel1"),
				Check:  resource.TestCheckResourceAttr("upcloud_server.this", "tags.#", "0"),
			},
		},
	})
}

func TestUpcloudServer_networkInterface(t *testing.T) {
	var serverID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerNetworkInterfaceConfig(
					networkInterface{
						niType: "utility",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "network_interface.#", "1"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.network"),
					testAccGetServerID("upcloud_server.this", &serverID),
					testAccCheckServerIDEqual("upcloud_server.this", &serverID),
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
						"upcloud_server.this", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.1.ip_address"),
					testAccCheckServerIDEqual("upcloud_server.this", &serverID),
					testAccCheckNetwork("upcloud_server.this", 1, "upcloud_network.test_network_1"),
					testAccGetServerID("upcloud_server.this", &serverID),
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
						"upcloud_server.this", "plan", "2xCPU-4GB"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this", "network_interface.#", "2"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.type",
						"utility"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.0.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.0.network"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.1.type",
						"private"),
					resource.TestCheckResourceAttr(
						"upcloud_server.this",
						"network_interface.1.ip_address_family",
						"IPv4"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.1.mac_address"),
					resource.TestCheckResourceAttrSet(
						"upcloud_server.this",
						"network_interface.1.ip_address"),
					testAccCheckServerIDEqual("upcloud_server.this", &serverID),
					testAccCheckNetwork("upcloud_server.this", 1, "upcloud_network.test_network_11"),
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

func testAccCheckServerIDEqual(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		newID := s.RootModule().Resources[resourceName].Primary.ID
		if newID != *id {
			return fmt.Errorf("new server ID unexpectedly does not equal old ID: %s == %s", newID, *id)
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

type networkInterface struct {
	niType     string
	network    bool
	newNetwork bool
}

func testAccServerNetworkInterfaceConfig(nis ...networkInterface) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
		resource "upcloud_server" "this" {
			zone     = "fi-hel1"
			hostname = "tf-acc-test-server-network-interface"
			plan     = "2xCPU-4GB"
			metadata = true

			template {
					storage = "%s"
					size = 10
			}
	`, debianTemplateUUID))

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
		if ni.network {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "tf-acc-test-server-network-interface-net-%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, i, i, i, i))
		}

		if ni.newNetwork {
			builder.WriteString(fmt.Sprintf(`
				resource "upcloud_network" "test_network_%d" {
					name = "tf-acc-test-server-network-interface-net-%d"
					zone = "fi-hel1"

					ip_network {
						address = "10.0.%d.0/24"
						dhcp = true
						dhcp_default_route = false
						family = "IPv4"
						gateway = "10.0.%d.1"
					}
				}
			`, 10+i, 10+i, 10+i, 10+i))
		}
	}

	return builder.String()
}

func TestUpcloudServer_updatePreChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configSimple("tf-acc-test-server-update-pre-checks", "1xCPU-1GB", "fi-hel2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("upcloud_server.this", "plan"),
				),
			},
			{
				// Test updating with invalid plan
				Config:             configSimple("tf-acc-test-server-create-pre-checks", "1xCPU-1G", "fi-hel1"),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("expected plan to be one of"),
				Check:              resource.TestCheckResourceAttr("upcloud_server.this", "plan", "1xCPU-1GB"),
			},
		},
	})
}

func TestUpcloudServer_createPreChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test creating with invalid plan
				Config:             configSimple("tf-acc-test-server-create-pre-checks", "1xCPU-1G", "fi-hel1"),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("expected plan to be one of"),
			},
			{
				// Test creating with invalid zone
				Config:             configSimple("tf-acc-test-server-create-pre-checks", "1xCPU-1GB", "_fi-hel2"),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("expected zone to be one of"),
			},
		},
	})
}

func configHotResize(planName string, hotResize bool, captureUptime bool, checkUptime bool, keyDir string) string {
	provisioner := ""

	if captureUptime {
		provisioner = `
			provisioner "remote-exec" {
				inline = [
					"uptime -s > /tmp/server_start_time.txt",
				]
				connection {
					type        = "ssh"
					user        = "root"
					host        = self.network_interface[0].ip_address
					private_key = file("%s/id_rsa")
				}
			}
		`
		provisioner = fmt.Sprintf(provisioner, keyDir)
	} else if checkUptime {
		provisioner = `
			provisioner "remote-exec" {
				inline = [
					"if [ -f /tmp/server_start_time.txt ]; then",
					"  ORIGINAL_START_TIME=$(cat /tmp/server_start_time.txt)",
					"  CURRENT_START_TIME=$(uptime -s)",
					"  echo \"Original start time: $ORIGINAL_START_TIME\"",
					"  echo \"Current start time: $CURRENT_START_TIME\"",
					"  if [ \"$ORIGINAL_START_TIME\" = \"$CURRENT_START_TIME\" ]; then",
					"    echo 'SUCCESS: Server was not restarted after hot resize'",
					"    exit 0",
					"  else",
					"    echo 'ERROR: Server was restarted after hot resize'",
					"    exit 1",
					"  fi",
					"else",
					"  echo 'ERROR: Could not find server start time file'",
					"  exit 1",
					"fi",
				]
				connection {
					type        = "ssh"
					user        = "root"
					host        = self.network_interface[0].ip_address
					private_key = file("%s/id_rsa")
				}
			}
		`
		provisioner = fmt.Sprintf(provisioner, keyDir)
	}

	return fmt.Sprintf(`
		resource "upcloud_server" "hot_resize" {
			hostname    = "tf-acc-test-server-hot-resize"
			zone        = "pl-waw1"
			plan        = "%s"
			metadata    = true
			hot_resize  = %t

			login {
				user = "root"
				keys = [
					file("%s/id_rsa.pub")
				]
			}

			template {
				storage = "%s"
				size    = 10
			}

			network_interface {
				type = "public"
			}

			%s
		}

		output "server_ip" {
			value = upcloud_server.hot_resize.network_interface[0].ip_address
		}
	`, planName, hotResize, keyDir, debianTemplateUUID, provisioner)
}

// generateSSHKey generates an SSH key pair in the given directory
func generateSSHKey(t *testing.T, keyDir string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create private key file
	privateKeyFile := filepath.Join(keyDir, "id_rsa")
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	if err := os.WriteFile(privateKeyFile, privateKeyPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Create public key file
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pub)
	publicKeyFile := filepath.Join(keyDir, "id_rsa.pub")
	if err := os.WriteFile(publicKeyFile, publicKeyBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	t.Logf("Temporary SSH keys generated successfully in %s", keyDir)
	return nil
}

func TestUpcloudServer_hotResize(t *testing.T) {
	// Skip if we're not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping hot resize test as TF_ACC is not set")
	}

	// Create a temporary directory for SSH keys
	keyDir := t.TempDir()
	err := generateSSHKey(t, keyDir)
	if err != nil {
		t.Fatalf("Failed to generate SSH keys: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Create a server with 1xCPU-1GB plan and capture uptime
				Config: configHotResize("1xCPU-1GB", true, true, false, keyDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.hot_resize", "plan", "1xCPU-1GB"),
					resource.TestCheckResourceAttr("upcloud_server.hot_resize", "hot_resize", "true"),
					func(_ *terraform.State) error {
						t.Logf("Initial server startup and uptime capture complete")
						return nil
					},
				),
			},
			{
				// Step 2: Apply hot resize to 1xCPU-2GB
				Config: configHotResize("1xCPU-2GB", true, false, false, keyDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.hot_resize", "plan", "1xCPU-2GB"),
					resource.TestCheckResourceAttr("upcloud_server.hot_resize", "hot_resize", "true"),
					func(_ *terraform.State) error {
						t.Logf("Hot resize applied, server plan changed to 1xCPU-2GB")
						return nil
					},
				),
			},
			{
				// Step 3: Verify that the server didn't restart by checking the uptime in a separate step
				Config: configHotResize("1xCPU-2GB", true, false, true, keyDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.hot_resize", "plan", "1xCPU-2GB"),
					func(_ *terraform.State) error {
						t.Logf("Server was successfully hot resize'd")
						return nil
					},
				),
			},
		},
	})
}

func TestUpcloudServer_hotResizeWithNetworkChange(t *testing.T) {
	// Skip if we're not running acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping hot resize with network change test as TF_ACC is not set")
	}

	// Create a temporary directory for SSH keys
	keyDir := t.TempDir()
	err := generateSSHKey(t, keyDir)
	if err != nil {
		t.Fatalf("Failed to generate SSH keys: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Create a server with 1xCPU-1GB plan and hot_resize=true, and capture uptime
				Config: fmt.Sprintf(`
					resource "upcloud_server" "mixed_changes" {
						hostname    = "tf-acc-test-server-mixed-changes"
						zone        = "pl-waw1"
						plan        = "1xCPU-1GB"
						metadata    = true
						hot_resize  = true

						login {
							user = "root"
							keys = [
								file("%s/id_rsa.pub")
							]
						}

						template {
							storage = "%s"
							size    = 10
						}

						network_interface {
							type = "public"
						}

						# Capture the initial server uptime for comparison later
						provisioner "remote-exec" {
							inline = [
								"uptime -s > /tmp/server_start_time.txt",
								"echo 'Initial server start time captured'",
							]
							connection {
								type        = "ssh"
								user        = "root"
								host        = self.network_interface[0].ip_address
								private_key = file("%s/id_rsa")
							}
						}
					}

					output "server_ip" {
						value = upcloud_server.mixed_changes.network_interface[0].ip_address
					}
				`, keyDir, debianTemplateUUID, keyDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "plan", "1xCPU-1GB"),
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "hot_resize", "true"),
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "network_interface.#", "1"),
				),
			},
			{
				// Step 2: Attempt to make both hot-resize-compatible change (plan) and a change requiring server restart (add network)
				Config: fmt.Sprintf(`
					resource "upcloud_server" "mixed_changes" {
						hostname    = "tf-acc-test-server-mixed-changes"
						zone        = "pl-waw1"
						plan        = "1xCPU-2GB"  # Changed plan
						metadata    = true
						hot_resize  = true

						login {
							user = "root"
							keys = [
								file("%s/id_rsa.pub")
							]
						}

						template {
							storage = "%s"
							size    = 10
						}

						network_interface {
							type = "public"
						}

						network_interface {
							type = "utility"  # Added network interface
						}

						# Check if the server was restarted by comparing the uptime
						provisioner "remote-exec" {
							inline = [
								"if [ -f /tmp/server_start_time.txt ]; then",
								"  ORIGINAL_START_TIME=$(cat /tmp/server_start_time.txt)",
								"  CURRENT_START_TIME=$(uptime -s)",
								"  echo \"Original start time: $ORIGINAL_START_TIME\"",
								"  echo \"Current start time: $CURRENT_START_TIME\"",
								"  if [ \"$ORIGINAL_START_TIME\" = \"$CURRENT_START_TIME\" ]; then",
								"    echo 'ERROR: Server was not restarted when it should have been'",
								"    exit 1",
								"  else",
								"    echo 'Server was correctly restarted as expected'",
								"  fi",
								"else",
								"  echo 'ERROR: Could not find server start time file'",
								"  exit 1",
								"fi",
							]
							connection {
								type        = "ssh"
								user        = "root"
								host        = self.network_interface[0].ip_address
								private_key = file("%s/id_rsa")
							}
						}
					}

					output "server_ip" {
						value = upcloud_server.mixed_changes.network_interface[0].ip_address
					}
				`, keyDir, debianTemplateUUID, keyDir),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "plan", "1xCPU-2GB"),
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "hot_resize", "true"),
					resource.TestCheckResourceAttr("upcloud_server.mixed_changes", "network_interface.#", "2"),
					func(_ *terraform.State) error {
						t.Logf("Successfully applied both plan change and network interface change")
						t.Logf("Server was restarted as expected when both hot resize and network changes were applied")
						return nil
					},
				),
			},
		},
	})
}
