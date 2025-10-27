package objectstorage

import (
	"regexp"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUpcloudManagedObjectStorage(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_s2.tf")

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
					resource.TestCheckResourceAttr(this, "region", "europe-1"),
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
	testDataE := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_e.tf")

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
			errorRe: regexp.MustCompile(`must only contain printable ASCII characters and must not start with`),
		},
		{
			labels:  `_key = "key-starts-with-underscore"`,
			errorRe: regexp.MustCompile(`must only contain printable ASCII characters and must not start with`),
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
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps:                    steps,
	})
}

func TestAccUpcloudManagedObjectStorage_CustomDomain(t *testing.T) {
	// The test does not configure the required DNS settings for the custom domain to work. This will cause the object storage instance to be stuck in a pending state and thus it cannot be modified as any modification will cause the provider to wait until the instance reaches running state.
	testDataS1 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_custom_domain_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_custom_domain_s2.tf")
	testDataS3 := utils.ReadTestDataFile(t, "../testdata/upcloud_managed_object_storage/managed_object_storage_custom_domain_s3.tf")

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
			},
			{
				Config: testDataS3,
			},
		},
	})
}
