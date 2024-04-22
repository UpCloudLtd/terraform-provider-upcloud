package upcloud

import (
	"regexp"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUpcloudManagedObjectStorage(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_s2.tf")

	this := "upcloud_managed_object_storage.this"
	minimal := "upcloud_managed_object_storage.minimal"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(minimal, "name", "tf-acc-test-objstov2-minimal"),
					resource.TestCheckResourceAttr(this, "name", "tf-acc-test-objstov2-complex"),
					resource.TestCheckResourceAttr(this, "region", "europe-1"),
					resource.TestCheckResourceAttr(this, "configured_status", "started"),
					resource.TestCheckResourceAttr(this, "labels.%", "2"),
					resource.TestCheckResourceAttr(this, "labels.test", "objsto2-tf"),
					resource.TestCheckResourceAttr(this, "network.#", "2"),
				),
			},
			{
				Config:            testDataS2,
				ImportStateVerify: true,
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
	testDataE := utils.ReadTestDataFile(t, "testdata/upcloud_managed_object_storage/managed_object_storage_e.tf")

	labelsPlaceholder := `TEST_KEY = "TEST_VALUE"`
	stepsData := []struct {
		labels  string
		errorRe *regexp.Regexp
	}{
		{
			labels:  `t = "too-short-key"`,
			errorRe: regexp.MustCompile(`Map key lengths should be in the range \(2 - 32\)`),
		},
		{
			labels:  `test-validation-fails-if-label-name-too-long = ""`,
			errorRe: regexp.MustCompile(`Map key lengths should be in the range \(2 - 32\)`),
		},
		{
			labels:  `test-validation-fails-åäö = "invalid-characters-in-key"`,
			errorRe: regexp.MustCompile(`Map key expected to match regular expression`),
		},
		{
			labels:  `test-validation-fails = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam egestas dolor vitae erat egestas, vel malesuada nisi ullamcorper. Aenean suscipit turpis quam, ut interdum lorem varius dignissim. Morbi eu erat bibendum, tincidunt turpis id, porta enim. Pellentesque..."`,
			errorRe: regexp.MustCompile(`Map value lengths should be in the range \(0 - 255\)`),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProviderFactories,
		Steps:                    steps,
	})
}
