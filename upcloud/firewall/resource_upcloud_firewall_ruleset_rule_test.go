package firewalltests

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUpCloudFirewallRuleset_withRules creates a ruleset with two rules, verifies
// order and positions, imports the ruleset (including rules), then updates the rule
// list by reordering and adding a third rule, and finally clears all rules.
func TestAccUpCloudFirewallRuleset_withRules(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rules_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rules_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rules_s3.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-rules-%s", acctest.RandString(10))
	vars := map[string]config.Variable{"ruleset_name": config.StringVariable(rulesetName)}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create ruleset with two rules.
				Config:          testDataS1,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.#", "2"),
					// First rule: HTTP
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.direction", "in"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.destination_port_start", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.destination_port_end", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.comment", "Allow HTTP"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.position", "1"),
					// Second rule: HTTPS
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.destination_port_start", "443"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.comment", "Allow HTTPS"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.position", "2"),
				),
			},
			{
				// Import the ruleset; rules must be read back correctly.
				ResourceName:      "upcloud_firewall_ruleset.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"updated_at", "version"},
				Config:            testDataS1,
				ConfigVariables:   vars,
			},
			{
				// Reorder (HTTPS first, HTTP second) and add SSH as third rule.
				Config:          testDataS2,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.#", "3"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.destination_port_start", "443"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.comment", "Allow HTTPS"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.0.position", "1"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.destination_port_start", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.comment", "Allow HTTP"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.1.position", "2"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.2.destination_port_start", "22"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.2.comment", "Allow SSH"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.2.position", "3"),
				),
			},
			{
				// Clear all rules.
				Config:          testDataS3,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.#", "0"),
				),
			},
		},
	})
}
