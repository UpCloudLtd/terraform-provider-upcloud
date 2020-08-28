package upcloud

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider
var testAccProviderFunc func() *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"upcloud": testAccProvider,
	}

	testAccProviderFactories = func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {

		var providerNames = []string{"upcloud"}
		var factories = make(map[string]func() (*schema.Provider, error), len(providerNames))
		for _, name := range providerNames {
			p := Provider()
			factories[name] = func() (*schema.Provider, error) { //nolint:unparam
				return p, nil
			}
			*providers = append(*providers, p)
		}
		return factories
	}
	testAccProviderFunc = func() *schema.Provider { return testAccProvider }
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_USERNAME"); v == "" {
		t.Fatal("UPCLOUD_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_PASSWORD must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
