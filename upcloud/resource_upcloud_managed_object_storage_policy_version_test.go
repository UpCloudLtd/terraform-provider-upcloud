package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccUpcloudManagedObjectStoragePolicy_Versioning(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t,
		"testdata/upcloud_managed_object_storage/managed_object_storage_policy_version_s1.tf",
	)
	testDataS2 := utils.ReadTestDataFile(t,
		"testdata/upcloud_managed_object_storage/managed_object_storage_policy_version_s2.tf",
	)

	policy := "upcloud_managed_object_storage_policy.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Create policy
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(policy, "name", "versioned-policy"),
					resource.TestCheckResourceAttrSet(policy, "default_version_id"),
				),
			},
			{
				// Step 2: Update document - check an update is planned
				Config: testDataS2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(policy, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(policy, "default_version_id"),
				),
			},
			{
				// Step 3: Ensure no drift
				Config:   testDataS2,
				PlanOnly: true,
			},
		},
	})
}
