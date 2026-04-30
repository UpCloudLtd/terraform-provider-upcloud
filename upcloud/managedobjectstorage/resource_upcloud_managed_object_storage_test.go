package managedobjectstoragetests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccUpcloudManagedObjectStorage(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_s2.tf")

	this := "upcloud_managed_object_storage.this"
	minimal := "upcloud_managed_object_storage.minimal"
	bucket := "upcloud_managed_object_storage_bucket.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(minimal, "name", "tf-acc-test-objstov2-minimal"),
					resource.TestCheckResourceAttr(this, "name", "tf-acc-test-objstov2-complex"),
					resource.TestCheckResourceAttr(this, "region", "europe-3"),
					resource.TestCheckResourceAttr(this, "configured_status", "started"),
					resource.TestCheckResourceAttr(this, "labels.%", "2"),
					resource.TestCheckResourceAttr(this, "labels.test", "objsto2-tf"),
					resource.TestCheckResourceAttr(this, "network.#", "2"),
				),
			},
			{
				Config:                  testDataS1,
				ResourceName:            this,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"operational_state", "updated_at"},
			},
			{
				Config:            testDataS1,
				ResourceName:      bucket,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:                  testDataS1,
				ResourceName:            minimal,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"operational_state", "updated_at"},
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(minimal, "name", "tf-acc-test-objstov2-renamed"),
					resource.TestCheckResourceAttr(this, "configured_status", "started"),
					resource.TestCheckResourceAttr(this, "labels.owned-by", "team-devex"),
					resource.TestCheckResourceAttr(this, "network.#", "1"),
				),
			},
		},
	})
}

func TestAccUpcloudManagedObjectStorage_LabelsValidation(t *testing.T) {
	testDataE := utils.ReadTestDataFile(t, "testdata/managed_object_storage_e.tf")

	labelsPlaceholder := `TEST_KEY = "TEST_VALUE"`
	stepsData := []struct {
		labels  string
		errorRe *regexp.Regexp
	}{
		{
			labels:  `t = "too-short-key"`,
			errorRe: upcloud.IgnoreWhitespaceDiff(`string length must be between 2 and 32`),
		},
		{
			labels:  `test-validation-fails-if-label-name-too-long = ""`,
			errorRe: upcloud.IgnoreWhitespaceDiff(`string length must be between 2 and 32`),
		},
		{
			labels:  `test-validation-fails-åäö = "invalid-characters-in-key"`,
			errorRe: upcloud.IgnoreWhitespaceDiff(`must only contain printable ASCII characters and must not start with`),
		},
		{
			labels:  `_key = "key-starts-with-underscore"`,
			errorRe: upcloud.IgnoreWhitespaceDiff(`must only contain printable ASCII characters and must not start with`),
		},
		{
			labels:  `test-validation-fails = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam egestas dolor vitae erat egestas, vel malesuada nisi ullamcorper. Aenean suscipit turpis quam, ut interdum lorem varius dignissim. Morbi eu erat bibendum, tincidunt turpis id, porta enim. Pellentesque..."`,
			errorRe: upcloud.IgnoreWhitespaceDiff(`string length must be between 0 and 255`),
		},
	}
	var steps []resource.TestStep
	for _, step := range stepsData {
		steps = append(steps, resource.TestStep{
			Config:      strings.Replace(testDataE, labelsPlaceholder, step.labels, 1),
			ExpectError: step.errorRe,
			PlanOnly:    true,
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps:                    steps,
	})
}

func TestAccUpcloudManagedObjectStorage_CustomDomain(t *testing.T) {
	// The test does not configure the required DNS settings for the custom domain to work. This will cause the object storage instance to be stuck in a pending state and thus it cannot be modified as any modification will cause the provider to wait until the instance reaches running state.
	testDataS1 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_custom_domain_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_custom_domain_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_custom_domain_s3.tf")

	objsto := "upcloud_managed_object_storage.this"
	customDomain := "upcloud_managed_object_storage_custom_domain.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(customDomain, "domain_name", "objects.example.com"),
				),
			},
			{
				Config:            testDataS1,
				ResourceName:      customDomain,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(customDomain, "domain_name", "obj.example.com"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(customDomain, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testDataS3,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(customDomain, plancheck.ResourceActionDestroy),
					},
				},
			},
		},
	})
}

// TestAccUpcloudManagedObjectStorage_FullNetworkReplace tests that a complete network replacement
// (no overlap between old and new set) succeeds. The two-step update sends an empty intermediate
// network list, which verifies the API accepts it before adding the new networks.
func TestAccUpcloudManagedObjectStorage_FullNetworkReplace(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_full_network_replace_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_full_network_replace_s2.tf")

	objsto := "upcloud_managed_object_storage.this"
	prefix := fmt.Sprintf("tf-acc-test-objsto-fullswap-%s-", acctest.RandString(4))
	variables := map[string]config.Variable{
		"prefix": config.StringVariable(prefix),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS1,
				ConfigVariables:          variables,
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS2,
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(objsto, "network.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS2,
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccUpcloudManagedObjectStorage_NetworkChange(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_network_change_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_network_change_s2.tf")

	objsto := "upcloud_managed_object_storage.this"
	prefix := fmt.Sprintf("tf-acc-test-objsto-swap-%s-", acctest.RandString(4))
	variables := map[string]config.Variable{
		"prefix": config.StringVariable(prefix),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { upcloud.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS1,
				ConfigVariables:          variables,
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS2,
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(objsto, "network.#", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
				Config:                   testDataS2,
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(objsto, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}
