package firewalltests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/credentials"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUpCloudServerFirewallRuleset_basic(t *testing.T) {
	serverName := fmt.Sprintf("tf-acc-test-server-ruleset-%s", acctest.RandString(10))
	networkName := fmt.Sprintf("tf-acc-test-server-ruleset-network-%s", acctest.RandString(10))
	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))
	resourceName := "upcloud_server_firewall_ruleset.test"
	var serverID, rulesetID string

	config := testAccServerFirewallRulesetConfig(serverName, networkName, rulesetName, true)
	configWithoutAttachment := testAccServerFirewallRulesetConfig(serverName, networkName, rulesetName, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "server_id", "upcloud_server.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ruleset_id", "upcloud_firewall_ruleset.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testAccCaptureServerFirewallRulesetIDs(resourceName, &serverID, &rulesetID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: configWithoutAttachment,
				Check:  testAccCheckPrivateFirewallRulesetDetached(&serverID, &rulesetID),
			},
		},
	})
}

func testAccServerFirewallRulesetConfig(serverName, networkName, rulesetName string, withAttachment bool) string {
	attachment := ""
	if withAttachment {
		attachment = `
		resource "upcloud_server_firewall_ruleset" "test" {
			server_id  = upcloud_server.test.id
			ruleset_id = upcloud_firewall_ruleset.test.id
		}`
	}

	return fmt.Sprintf(`
		resource "upcloud_network" "test" {
			name = "%s"
			zone = "fi-hel1"

			ip_network {
				address = "172.24.1.0/24"
				dhcp    = true
				family  = "IPv4"
			}
		}

		resource "upcloud_server" "test" {
			zone     = "fi-hel1"
			hostname = "%s"
			plan     = "1xCPU-1GB"
			metadata = true

			template {
				storage = "%s"
				size    = 10
			}

			network_interface {
				type    = "private"
				network = upcloud_network.test.id
			}
		}

		resource "upcloud_firewall_ruleset" "test" {
			name        = "%s"
			description = "Test private ruleset"

			rules = [
				{
					action                 = "accept"
					direction              = "in"
					family                 = "IPv4"
					protocol               = "tcp"
					comment                = "Allow SSH"
					destination_port_start = 22
					destination_port_end   = 22
				}
			]
		}

		%s
	`, networkName, serverName, upcloud.DebianTemplateUUID, rulesetName, attachment)
}

func testAccCaptureServerFirewallRulesetIDs(resourceName string, serverID, rulesetID *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		*serverID = rs.Primary.Attributes["server_id"]
		*rulesetID = rs.Primary.Attributes["ruleset_id"]
		return nil
	}
}

func testAccCheckPrivateFirewallRulesetDetached(serverID, rulesetID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		serverUUID, err := uuid.Parse(*serverID)
		if err != nil {
			return fmt.Errorf("parse server UUID: %w", err)
		}

		rulesetUUID, err := uuid.Parse(*rulesetID)
		if err != nil {
			return fmt.Errorf("parse ruleset UUID: %w", err)
		}

		creds, err := credentials.Parse(credentials.Credentials{})
		if err != nil {
			return fmt.Errorf("parse credentials: %w", err)
		}

		client, err := v9.New("", v9.WithCredentials(creds))
		if err != nil {
			return fmt.Errorf("create v9 client: %w", err)
		}

		apiResp, err := client.ListPrivateFirewallRulesetRelationshipsWithResponse(context.Background(), serverUUID)
		if err != nil {
			return fmt.Errorf("list private firewall ruleset relationships: %w", err)
		}
		if apiResp.StatusCode() != http.StatusOK {
			return fmt.Errorf("list private firewall ruleset relationships returned %s", apiResp.Status())
		}
		if apiResp.JSON200 == nil {
			return fmt.Errorf("list private firewall ruleset relationships returned no response body")
		}

		for _, relationship := range apiResp.JSON200.FirewallRulesetRelationships.Private {
			if relationship.FirewallRulesetUuid != nil && *relationship.FirewallRulesetUuid == rulesetUUID {
				return fmt.Errorf("private firewall ruleset %s is still attached to server %s", rulesetUUID, serverUUID)
			}
		}
		return nil
	}
}
