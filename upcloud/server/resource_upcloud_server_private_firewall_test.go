package servertests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccUpCloudServer_privateFirewall(t *testing.T) {
	name := "tf-acc-private-firewall-" + acctest.RandString(8)
	enabled := utils.ReadTestDataFile(t, "testdata/server_private_firewall_enabled.tf")
	disabled := utils.ReadTestDataFile(t, "testdata/server_private_firewall_disabled.tf")
	omitted := utils.ReadTestDataFile(t, "testdata/server_private_firewall_omitted.tf")
	variables := map[string]config.Variable{
		"name":     config.StringVariable(name),
		"template": config.StringVariable(upcloud.DebianTemplateUUID),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:          enabled,
				Check:           check("true", "drop", "accept"),
				ConfigVariables: variables,
			},
			{
				Config:          enabled,
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:      "upcloud_server.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:          disabled,
				Check:           check("false", "accept", "drop"),
				ConfigVariables: variables,
			},
			{
				Config:          enabled,
				Check:           check("true", "drop", "accept"),
				ConfigVariables: variables,
			},
			// Optional/computed attributes retain the observed values when omitted.
			{
				Config:          omitted,
				Check:           check("true", "drop", "accept"),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccUpCloudServer_privateFirewallCreate(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]string
	}{
		{
			name: "disabled",
			expected: map[string]string{
				"firewall_private":                         "false",
				"firewall_private_default_incoming_action": "accept",
				"firewall_private_default_outgoing_action": "drop",
			},
		},
		{name: "omitted"},
		{
			name: "actions_only",
			expected: map[string]string{
				"firewall_private_default_incoming_action": "drop",
				"firewall_private_default_outgoing_action": "accept",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := "tf-acc-private-firewall-" + acctest.RandString(8)
			testdata := utils.ReadTestDataFile(t, "testdata/server_private_firewall_"+tt.name+".tf")
			variables := map[string]config.Variable{
				"name":     config.StringVariable(name),
				"template": config.StringVariable(upcloud.DebianTemplateUUID),
			}
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("upcloud_server.test", "firewall", "true"),
				resource.TestCheckResourceAttr("data.upcloud_server.test", "firewall", "true"),
			}
			for _, attribute := range []string{"firewall_private", "firewall_private_default_incoming_action", "firewall_private_default_outgoing_action"} {
				checks = append(checks,
					resource.TestCheckResourceAttrSet("upcloud_server.test", attribute),
					resource.TestCheckResourceAttrPair("upcloud_server.test", attribute, "data.upcloud_server.test", attribute),
				)
			}
			for attribute, value := range tt.expected {
				checks = append(checks, resource.TestCheckResourceAttr("upcloud_server.test", attribute, value))
			}
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:          testdata,
						Check:           resource.ComposeAggregateTestCheckFunc(checks...),
						ConfigVariables: variables,
					},
					{
						Config:          testdata,
						ConfigVariables: variables,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectEmptyPlan(),
							},
						},
					},
					{
						ResourceName:      "upcloud_server.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}
