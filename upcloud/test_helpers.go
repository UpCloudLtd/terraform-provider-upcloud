package upcloud

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	TestAccProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	testAccProvider          *schema.Provider
)

const debianTemplateUUID = "01000000-0000-4000-8000-000020070100"

func init() {
	testAccProvider = Provider()
	TestAccProviderFactories = make(map[string]func() (tfprotov6.ProviderServer, error))
	TestAccProviderFactories["upcloud"] = func() (tfprotov6.ProviderServer, error) {
		factory, err := NewProviderServerFactory()
		return factory(), err
	}
}

func TestAccPreCheck(t *testing.T) {
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

func IgnoreWhitespaceDiff(str string) *regexp.Regexp {
	ws := regexp.MustCompile(`\s+`)
	re := ws.ReplaceAllString(str, `\s+`)
	return regexp.MustCompile(re)
}
