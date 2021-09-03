package upcloud

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	TestAccProviders         map[string]*schema.Provider
	TestAccProviderFactories func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error)
	TestAccProvider          *schema.Provider
	TestAccProviderFunc      func() *schema.Provider
)

func init() {
	TestAccProvider = Provider()
	TestAccProviders = map[string]*schema.Provider{
		"upcloud": TestAccProvider,
	}

	TestAccProviderFactories = func(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
		providerNames := []string{"upcloud"}
		factories := make(map[string]func() (*schema.Provider, error), len(providerNames))
		for _, name := range providerNames {
			p := Provider()
			factories[name] = func() (*schema.Provider, error) { //nolint:unparam
				return p, nil
			}
			*providers = append(*providers, p)
		}
		return factories
	}
	TestAccProviderFunc = func() *schema.Provider { return TestAccProvider }
}

func AccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("UPCLOUD_USERNAME"); v == "" {
		t.Fatal("UPCLOUD_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_PASSWORD must be set for acceptance tests")
	}

	err := TestAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
