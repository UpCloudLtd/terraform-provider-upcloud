package upcloud

import (
	"regexp"
	"strings"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudGateway(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/upcloud_gateway/gateway_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/upcloud_gateway/gateway_s2.tf")

	var providers []*schema.Provider
	name := "upcloud_gateway.this"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-acc-test-net-gateway-gw"),
					resource.TestCheckResourceAttr(name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(name, "features.0", "nat"),
					resource.TestCheckResourceAttr(name, "configured_status", "started"),
					resource.TestCheckResourceAttr(name, "labels.%", "2"),
					resource.TestCheckResourceAttr(name, "labels.test", "net-gateway-tf"),
					resource.TestCheckResourceAttr(name, "labels.owned-by", "team-iaas"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "addresses.*", map[string]string{"name": "public-ip-1"}),
				),
			},
			{
				Config:            testDataS2,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-acc-test-net-gateway-gw-renamed"),
					resource.TestCheckResourceAttr(name, "configured_status", "stopped"),
					resource.TestCheckResourceAttr(name, "labels.owned-by", "team-devex"),
				),
			},
		},
	})
}

func TestAccUpcloudGateway_LabelsValidation(t *testing.T) {
	testDataE := utils.ReadTestDataFile(t, "testdata/upcloud_gateway/gateway_e.tf")

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

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps:             steps,
	})
}
