package firewalltests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var providerv5491 = map[string]resource.ExternalProvider{
	"upcloud": {
		Source:            "UpCloudLtd/upcloud",
		VersionConstraint: "5.49.1",
	},
}

// TestAccUpCloudFirewallRuleset_upgradeFromV5_49_1 sets every optional field with v5.49.1.
// The same config must plan nothing with the current provider.
// Removing the fields must then clear them. The test framework fails the step if a removed value comes back on refresh.
func TestAccUpCloudFirewallRuleset_upgradeFromV5_49_1(t *testing.T) {
	allFields := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_all_fields.tf")
	allFieldsRemoved := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_all_fields_removed.tf")
	name := config.StringVariable("tf-acc-test-ruleset-upgrade-" + acctest.RandString(10))
	disabled := map[string]config.Variable{"ruleset_name": name, "enabled": config.BoolVariable(false)}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerv5491,
				Config:            allFields,
				ConfigVariables:   map[string]config.Variable{"ruleset_name": name},
			},
			{
				// The API might ignore enabled = false on create, so disable the ruleset with an update.
				ExternalProviders: providerv5491,
				Config:            allFields,
				ConfigVariables:   disabled,
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   allFields,
				ConfigVariables:          disabled,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.2.protocol", "all"),
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   allFieldsRemoved,
				ConfigVariables:          map[string]config.Variable{"ruleset_name": name},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("upcloud_firewall_ruleset.test", "description"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "enabled", "true"),
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "default_dns_rules_enabled", "false"),
					resource.TestCheckNoResourceAttr("upcloud_firewall_ruleset.test", "rules.0.icmp_type"),
					resource.TestCheckNoResourceAttr("upcloud_firewall_ruleset.test", "rules.1.source_port_start"),
					resource.TestCheckNoResourceAttr("upcloud_firewall_ruleset.test", "rules.2.protocol"),
				),
			},
		},
	})
}

// TestAccUpCloudFirewallRuleset_upgradeNoRulesFromV5_49_1 creates a ruleset without rules with v5.49.1.
// v5.49.1 saves an empty rules list, and the current provider defaults rules to an empty list, so the plan must be empty.
func TestAccUpCloudFirewallRuleset_upgradeNoRulesFromV5_49_1(t *testing.T) {
	minimal := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_minimal.tf")
	vars := map[string]config.Variable{
		"ruleset_name": config.StringVariable("tf-acc-test-ruleset-upgrade-norules-" + acctest.RandString(10)),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerv5491,
				Config:            minimal,
				ConfigVariables:   vars,
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   minimal,
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}
