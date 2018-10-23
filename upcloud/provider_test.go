package upcloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	_ "github.com/terraform-providers/terraform-provider-template/template"
	_ "github.com/terraform-providers/terraform-provider-tls/tls"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvidersWithTLS map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var testAccTemplateProvider *schema.Provider

func init() {
	if v := os.Getenv("TF_ACC"); v == "1" {
		/* TODO: Getting errors with this on Travis-CI
		// *"github.com/hashicorp/terraform/helper/schema".Provider does not implement "github.com/terraform-providers/terraform-provider-template/vendor/github.com/hashicorp/terraform/terraform".ResourceProvider (wrong type for Apply method)

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
		*/
	}
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

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}
