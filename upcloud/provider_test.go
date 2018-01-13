package upcloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-template/template"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvidersWithTLS map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var testAccTemplateProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccTemplateProvider = template.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"upcloud":  testAccProvider,
		"template": testAccTemplateProvider,
	}
	testAccProvidersWithTLS = map[string]terraform.ResourceProvider{
		"tls": tls.Provider(),
	}

	for k, v := range testAccProviders {
		testAccProvidersWithTLS[k] = v
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_USERNAME"); v == "" {
		t.Fatal("UPCLOUD_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_PASSWORD must be set for acceptance tests")
	}

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}
