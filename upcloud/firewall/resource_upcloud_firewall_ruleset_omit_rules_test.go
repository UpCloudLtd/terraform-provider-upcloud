package firewalltests

import (
	"context"
	"net/http"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/credentials"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccUpCloudFirewallRuleset_omitRulesClearsRules removes the rules block after a rule is set.
// The removal must clear every rule in the ruleset, and the next plan must be empty.
// Then a rule added outside Terraform must also be planned for removal and cleared.
func TestAccUpCloudFirewallRuleset_omitRulesClearsRules(t *testing.T) {
	t.Parallel()

	rulesetName := "tf-acc-test-ruleset-norules-" + acctest.RandString(10)
	withRules := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_all_fields.tf")
	noRules := utils.ReadTestDataFile(t, "testdata/upcloud_firewall_ruleset_minimal.tf")
	vars := map[string]config.Variable{
		"ruleset_name": config.StringVariable(rulesetName),
	}
	var rulesetID string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   withRules,
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   noRules,
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.#", "0"),
					func(s *terraform.State) error {
						rulesetID = s.RootModule().Resources["upcloud_firewall_ruleset.test"].Primary.ID
						return nil
					},
				),
			},
			{
				// Add a rule outside Terraform. Refresh must show it, and apply must remove it.
				PreConfig:                func() { addRuleOutsideTerraform(t, rulesetID) },
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   noRules,
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("upcloud_firewall_ruleset.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("upcloud_firewall_ruleset.test", "rules.#", "0"),
			},
		},
	})
}

// addRuleOutsideTerraform adds one rule to the ruleset with the API client.
func addRuleOutsideTerraform(t *testing.T, rulesetID string) {
	t.Helper()

	rulesetUUID, err := uuid.Parse(rulesetID)
	if err != nil {
		t.Fatalf("parse ruleset UUID %q: %v", rulesetID, err)
	}
	creds, err := credentials.Parse(credentials.Credentials{})
	if err != nil {
		t.Fatalf("parse credentials: %v", err)
	}
	client, err := v9.New("", v9.WithCredentials(creds))
	if err != nil {
		t.Fatalf("create v9 client: %v", err)
	}

	resp, err := client.CreateFirewallRulesetRuleWithResponse(context.Background(), rulesetUUID, v9.CreateFirewallRulesetRuleJSONRequestBody{
		Action:    "accept",
		Direction: "in",
		Family:    "IPv4",
	})
	if err != nil {
		t.Fatalf("add rule: %v", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		t.Fatalf("add rule returned %s: %s", resp.Status(), resp.Body)
	}
}
