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

func TestAccUpCloudFirewallRulesetRule_basic(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_s1.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "direction", "in"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "family", "IPv4"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "protocol", "tcp"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_start", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_end", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "enabled", "true"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "comment", "Allow HTTP traffic"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset_rule.allow_http", "id"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset_rule.allow_http", "rule_id"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset_rule.allow_http", "ruleset_uuid"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.allow_http",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_update(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_s3.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_start", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_end", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "comment", "Allow HTTP traffic"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "enabled", "true"),
				),
			},
			{
				// Update rule - change action, ports, comment
				Config: testDataS2,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "action", "drop"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_start", "8080"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_end", "8080"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "comment", "Drop alternative HTTP port"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "enabled", "false"),
				),
			},
			{
				// Remove optional comment field - API doesn't clear it, field is sticky (PATCH semantics)
				Config: testDataS3,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// API keeps the previous comment value when it's removed from config
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "comment", "Drop alternative HTTP port"),
				),
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_addressRanges(t *testing.T) {
	testDataCIDR := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_cidr.tf")
	testDataRange := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_range.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test with CIDR notation
				Config: testDataCIDR,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_address_cidr", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "destination_address_cidr", "10.0.0.0/8"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testDataRange,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
			{
				// Test with IP range notation
				Config: testDataRange,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_address_start", "192.168.1.10"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_address_end", "192.168.1.20"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "destination_address_start", "10.0.0.1"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "destination_address_end", "10.0.0.254"),
				),
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_icmp(t *testing.T) {
	testData := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_icmp.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "protocol", "icmp"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "icmp_type", "8"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "family", "IPv4"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_multipleRules(t *testing.T) {
	t.Skip("API returns 500 Internal Server Error when creating multiple rules in parallel - known API issue under investigation")

	testData := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_multiple.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// HTTP rule
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "destination_port_start", "80"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_http", "position", "1"),
					// HTTPS rule
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_https", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_https", "destination_port_start", "443"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_https", "position", "2"),
					// SSH rule
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_ssh", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_ssh", "destination_port_start", "22"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.allow_ssh", "position", "3"),
					// Drop all rule
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.drop_all", "action", "drop"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.drop_all", "position", "100"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.allow_http",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.allow_https",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.allow_ssh",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.drop_all",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_ipv6(t *testing.T) {
	testData := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_ipv6.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "family", "IPv6"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_address_cidr", "2001:db8::/32"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "action", "accept"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_portRanges(t *testing.T) {
	testData := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_port_ranges.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_port_start", "1024"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "source_port_end", "65535"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "destination_port_start", "8000"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "destination_port_end", "9000"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}

func TestAccUpCloudFirewallRulesetRule_minimal(t *testing.T) {
	testData := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_rule_minimal.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "action", "accept"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "direction", "in"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset_rule.test", "family", "IPv4"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset_rule.test", "id"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset_rule.test", "rule_id"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testData,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
			},
		},
	})
}
