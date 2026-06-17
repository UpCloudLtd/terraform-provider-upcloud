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

func TestAccUpCloudFirewallRuleset_basic(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_s1.tf")

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
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetName),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "description", "Test firewall ruleset"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "true"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "default_dns_rules_enabled", "false"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "id"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "version"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "created_at"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "updated_at"),
				),
			},
			{
				Config: testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				ResourceName:      "upcloud_firewall_ruleset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUpCloudFirewallRuleset_update(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_s3.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-%s", acctest.RandString(10))
	rulesetNameUpdated := fmt.Sprintf("tf-acc-test-ruleset-updated-%s", acctest.RandString(10))

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
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetName),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "description", "Test firewall ruleset"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "true"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "default_dns_rules_enabled", "false"),
				),
			},
			{
				Config: testDataS2,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetNameUpdated),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetNameUpdated),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "false"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "default_dns_rules_enabled", "true"),
				),
			},
			{
				// Test removing optional description field - API doesn't clear it, field is sticky
				Config: testDataS3,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetNameUpdated),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetNameUpdated),
					// API keeps the previous value when description is removed from config
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccUpCloudFirewallRuleset_labels(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_labels_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_labels_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_labels_s3.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-labels-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with labels
				Config: testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetName),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.%", "3"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.env", "test"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.managed-by", "terraform"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.purpose", "acceptance-test"),
				),
			},
			{
				Config: testDataS1,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				ResourceName:      "upcloud_firewall_ruleset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update labels - modify and remove some
				Config: testDataS2,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.%", "2"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.env", "production"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.managed-by", "terraform"),
				),
			},
			{
				// Remove all labels
				Config: testDataS3,
				ConfigVariables: map[string]config.Variable{
					"ruleset_name": config.StringVariable(rulesetName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "labels.%", "0"),
				),
			},
		},
	})
}

// TestAccUpCloudFirewallRuleset_minimalToFull creates with only a name, imports,
// then adds all optional fields — covering both the minimal and null-to-value cases
// in one create/destroy cycle.
func TestAccUpCloudFirewallRuleset_minimalToFull(t *testing.T) {
	testDataMinimal := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_minimal.tf")
	testDataFull := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_s1.tf")

	rulesetName := fmt.Sprintf("tf-acc-test-ruleset-minimal-%s", acctest.RandString(10))
	vars := map[string]config.Variable{"ruleset_name": config.StringVariable(rulesetName)}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with only a name — verifies minimal config works and computed attrs are set.
				Config:          testDataMinimal,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetName),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "id"),
					resource.TestCheckResourceAttrSet("upcloud_firewall_ruleset.test", "version"),
					// Description is Optional+Computed; when omitted it must be null in state.
					resource.TestCheckNoResourceAttr("upcloud_firewall_ruleset.test", "description"),
				),
			},
			{
				ResourceName:      "upcloud_firewall_ruleset.test",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testDataMinimal,
				ConfigVariables:   vars,
			},
			{
				// Add all optional fields — verifies null-to-value transitions produce no spurious replace.
				Config:          testDataFull,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "name", rulesetName),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "description", "Test firewall ruleset"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "true"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "default_dns_rules_enabled", "false"),
				),
			},
		},
	})
}
