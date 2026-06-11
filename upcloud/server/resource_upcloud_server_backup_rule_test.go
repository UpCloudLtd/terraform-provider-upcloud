package servertests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUpCloudServer_removeBackupRule tests the scenario where a backup_rule
// is removed from a template after the server is created.
func TestAccUpCloudServer_removeBackupRule(t *testing.T) {
	s1 := utils.ReadTestDataFile(t, "testdata/server_backup_rule_s1.tf")
	s2 := utils.ReadTestDataFile(t, "testdata/server_backup_rule_s2.tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Create server with backup_rule in template
				Config: s1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.test_backup_rule", "template.0.backup_rule.0.interval", "daily"),
					resource.TestCheckResourceAttr("upcloud_server.test_backup_rule", "template.0.backup_rule.0.time", "0100"),
					resource.TestCheckResourceAttr("upcloud_server.test_backup_rule", "template.0.backup_rule.0.retention", "8"),
				),
			},
			{
				// Step 2: Remove backup_rule from template
				Config: s2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("upcloud_server.test_backup_rule", "template.0.backup_rule.#", "0"),
				),
			},
		},
	})
}
