package servertests

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccUpCloudServer_privateFirewall(t *testing.T) {
	name := "tf-acc-private-firewall-" + acctest.RandString(8)
	config := func(settings string) string {
		return fmt.Sprintf(`
resource "upcloud_network" "test" {
 name = %[1]q
 zone = "fi-hel1"
 ip_network { address = "172.24.1.0/24"
  dhcp = true
  family = "IPv4"
 }
}
# Separate storage keeps import verification independent of template clone provenance.
resource "upcloud_storage" "test" {
 title = %[1]q
 zone = "fi-hel1"
 size = 10
 tier = "maxiops"
 clone { id = %[2]q }
}
resource "upcloud_server" "test" {
 hostname = %[1]q
 zone = "fi-hel1"
 plan = "1xCPU-1GB"
 metadata = true
 firewall = true
 %[3]s
 storage_devices { storage = upcloud_storage.test.id }
 network_interface {
  type = "private"
  network = upcloud_network.test.id
 }
}
data "upcloud_server" "test" { id = upcloud_server.test.id }
`, name, upcloud.DebianTemplateUUID, settings)
	}
	check := func(enabled, incoming, outgoing string) resource.TestCheckFunc {
		checks := []resource.TestCheckFunc{resource.TestCheckResourceAttr("upcloud_server.test", "firewall", "true")}
		for attribute, value := range map[string]string{
			"firewall_private":                         enabled,
			"firewall_private_default_incoming_action": incoming,
			"firewall_private_default_outgoing_action": outgoing,
		} {
			checks = append(checks, resource.TestCheckResourceAttr("upcloud_server.test", attribute, value), resource.TestCheckResourceAttr("data.upcloud_server.test", attribute, value))
		}
		return resource.ComposeAggregateTestCheckFunc(checks...)
	}
	enabled := config(`firewall_private = true
 firewall_private_default_incoming_action = "drop"
 firewall_private_default_outgoing_action = "accept"`)
	disabled := config(`firewall_private = false
 firewall_private_default_incoming_action = "accept"
 firewall_private_default_outgoing_action = "drop"`)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{Config: enabled, Check: check("true", "drop", "accept")},
			{Config: enabled, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{ResourceName: "upcloud_server.test", ImportState: true, ImportStateVerify: true},
			{Config: disabled, Check: check("false", "accept", "drop")},
			{Config: enabled, Check: check("true", "drop", "accept")},
			// Optional/computed attributes retain the observed values when omitted.
			{Config: config(""), Check: check("true", "drop", "accept"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		},
	})
}
