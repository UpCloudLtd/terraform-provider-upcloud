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
	conn1Name := "upcloud_gateway_connection.this"
	conn2Name := "upcloud_gateway_connection.this2"
	tunnel1Name := "upcloud_gateway_connection_tunnel.this"
	tunnel2Name := "upcloud_gateway_connection_tunnel.this2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-acc-test-net-gateway-gw"),
					resource.TestCheckResourceAttr(name, "zone", "pl-waw1"),
					resource.TestCheckResourceAttr(name, "plan", "advanced"),
					resource.TestCheckResourceAttr(name, "features.0", "nat"),
					resource.TestCheckResourceAttr(name, "configured_status", "started"),
					resource.TestCheckResourceAttr(name, "labels.%", "2"),
					resource.TestCheckResourceAttr(name, "labels.test", "net-gateway-tf"),
					resource.TestCheckResourceAttr(name, "labels.owned-by", "team-iaas"),
					resource.TestCheckResourceAttr(name, "address.#", "1"),
					resource.TestCheckResourceAttr(name, "address.0.%", "2"),
					resource.TestCheckResourceAttr(name, "address.0.name", "my-public-ip"),
					resource.TestCheckResourceAttrSet(name, "address.0.address"),

					resource.TestCheckResourceAttr(conn1Name, "name", "test-connection"),
					resource.TestCheckResourceAttrSet(conn1Name, "gateway"),
					resource.TestCheckResourceAttr(conn1Name, "type", "ipsec"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.#", "1"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.name", "local-route"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.static_network", "10.123.123.0/24"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.#", "1"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.name", "remote-route"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.static_network", "100.123.123.0/24"),

					resource.TestCheckResourceAttr(tunnel1Name, "name", "test-tunnel"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "connection_id"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "operational_state"),
					resource.TestCheckResourceAttr(tunnel1Name, "local_address_name", "my-public-ip"),
					resource.TestCheckResourceAttr(tunnel1Name, "remote_address", "100.123.123.10"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_auth_psk.0.psk", "presharedkey1"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.child_rekey_time"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.dpd_delay"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.dpd_timeout"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.ike_lifetime"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.rekey_time"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase1_algorithms.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase1_dh_group_numbers.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase1_integrity_algorithms.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_algorithms.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_dh_group_numbers.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_integrity_algorithms.0"),

					resource.TestCheckResourceAttr(conn2Name, "name", "test-connection2"),
					resource.TestCheckResourceAttrSet(conn2Name, "gateway"),
					resource.TestCheckResourceAttr(conn2Name, "type", "ipsec"),
					resource.TestCheckResourceAttr(conn2Name, "local_route.#", "1"),
					resource.TestCheckResourceAttr(conn2Name, "local_route.0.name", "local-route2"),
					resource.TestCheckResourceAttr(conn2Name, "local_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn2Name, "local_route.0.static_network", "22.123.123.0/24"),
					resource.TestCheckResourceAttr(conn2Name, "remote_route.#", "1"),
					resource.TestCheckResourceAttr(conn2Name, "remote_route.0.name", "remote-route2"),
					resource.TestCheckResourceAttr(conn2Name, "remote_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn2Name, "remote_route.0.static_network", "222.123.123.0/24"),

					resource.TestCheckResourceAttr(tunnel2Name, "name", "test-tunnel2"), // Single check for tunnel2 just to make sure it gets created

					// This field is deprecated, can be removed later
					resource.TestCheckTypeSetElemNestedAttrs(name, "addresses.*", map[string]string{"name": "my-public-ip"}),
				),
			},
			{
				// Check that computed fields are updated properly after refresh
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "connections.#", "2"),
					resource.TestCheckResourceAttr(conn1Name, "tunnels.#", "2"),
				),
			},
			{
				Config:             testDataS1,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testDataS2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-acc-test-net-gateway-gw-renamed"),
					resource.TestCheckResourceAttr(name, "configured_status", "stopped"),
					resource.TestCheckResourceAttr(name, "labels.test", "net-gateway-tf"),
					resource.TestCheckResourceAttr(name, "labels.owned-by", "team-devex"),

					resource.TestCheckResourceAttr(conn1Name, "local_route.#", "1"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.name", "local-route-updated"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn1Name, "local_route.0.static_network", "11.123.123.0/24"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.#", "1"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.name", "remote-route-updated"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.type", "static"),
					resource.TestCheckResourceAttr(conn1Name, "remote_route.0.static_network", "111.123.123.0/24"),

					resource.TestCheckResourceAttr(tunnel1Name, "remote_address", "100.123.123.20"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_auth_psk.0.psk", "presharedkey2"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_properties.0.child_rekey_time", "1000"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_properties.0.rekey_time", "1000"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_properties.0.phase1_algorithms.#", "2"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_properties.0.phase1_algorithms.0", "aes128"),
					resource.TestCheckResourceAttr(tunnel1Name, "ipsec_properties.0.phase1_algorithms.1", "aes256"),
					// We also check all the other ipsec_properties of the tunnel just to make sure that update didn't remove them
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.dpd_delay"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.dpd_timeout"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.ike_lifetime"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase1_dh_group_numbers.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase1_integrity_algorithms.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_algorithms.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_dh_group_numbers.0"),
					resource.TestCheckResourceAttrSet(tunnel1Name, "ipsec_properties.0.phase2_integrity_algorithms.0"),
				),
			},
			{
				// Check that computed fields are updated properly after refresh
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "connections.#", "1"),
					resource.TestCheckResourceAttr(conn1Name, "tunnels.#", "1"),
				),
			},
			{
				Config:             testDataS2,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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
