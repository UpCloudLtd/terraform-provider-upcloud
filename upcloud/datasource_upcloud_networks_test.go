package upcloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUpCloudNetworksNoZone(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksConfig("", ""),
				Check:  testAccNetworks("data.upcloud_networks.my_example_networks", "", ""),
			},
		},
	})
}

func TestAccUpCloudNetworksWithZone(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksConfig("fi-hel1", ""),
				Check:  testAccNetworks("data.upcloud_networks.my_example_networks", "fi-hel1", ""),
			},
		},
	})
}

func TestAccUpCloudNetworksWithFilter(t *testing.T) {
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworksConfig("", "^Public.*"),
				Check:  testAccNetworks("data.upcloud_networks.my_example_networks", "", "^Public.*"),
			},
		},
	})
}

func testAccNetworksConfig(zone string, filterName string) string {
	builder := strings.Builder{}

	builder.WriteString(`data "upcloud_networks" "my_example_networks" {`)
	if zone != "" {
		builder.WriteString(fmt.Sprintf(`  zone = "%s"`, zone))
	}

	if filterName != "" {
		builder.WriteString(fmt.Sprintf(`  filter_name = "%s"`, filterName))
	}
	builder.WriteString("}")

	s := builder.String()
	fmt.Println(s)

	return s
}

func testAccNetworks(resourceName string, zone string, filterName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check that the expected resource exists
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		// Check that it has an ID in a date time format
		_, err := time.Parse(time.RFC3339Nano, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ID (%s) is not in expected format: %w", rs.Primary.ID, err)
		}

		a := rs.Primary.Attributes

		// Check that the networks list is of an appropriate size
		var networksSize int
		if networksSize, err = strconv.Atoi(a["networks.#"]); err != nil {
			return fmt.Errorf("unable to get networksSize: %w", err)
		}

		if networksSize < 1 {
			return fmt.Errorf("number of networks seems low: %d", networksSize)
		}

		// For each network perform checks
		for i := 0; i < networksSize; i++ {
			// Check that these fields aren't empty
			for _, key := range []string{"name", "type", "id", "zone"} {
				fullkey := fmt.Sprintf("networks.%d.%s", i, key)
				val := a[fullkey]
				if val == "" {
					return fmt.Errorf("%s is unexpectedly empty", fullkey)
				}
			}

			if filterName != "" {
				// Checked names return match filter
				name := a[fmt.Sprintf("networks.%d.name", i)]
				m, err := regexp.MatchString(filterName, name)
				if err != nil {
					return err
				}

				if !m {
					return fmt.Errorf("name %s does not match pattern %s", name, filterName)
				}
			}

			// Check that the ip_network are of an appropriate size
			var ipNetworksSize int
			fullkey := fmt.Sprintf("networks.%d.ip_network", i)
			if ipNetworksSize, err = strconv.Atoi(a[fullkey+".#"]); err != nil {
				return fmt.Errorf("unable to get ipNetworksSize (%s): %w", fullkey, err)
			}

			if ipNetworksSize < 1 {
				return fmt.Errorf("number ip_network seems low for %s", fullkey)
			}

			// For each ip_network perform checks.
			for j := 0; j < ipNetworksSize; j++ {
				// Check that these fields are not empty
				for _, key := range []string{"address", "family"} {
					ipnFullkey := fmt.Sprintf("%s.%d.%s", fullkey, j, key)
					val := a[ipnFullkey]
					if val == "" {
						return fmt.Errorf("%s is unexpectedly empty", ipnFullkey)
					}
				}
			}
		}

		return nil
	}
}
