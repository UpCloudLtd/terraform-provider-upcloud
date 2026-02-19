package upcloud

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tftest "github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	TestAccProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	TestAccProvider          *schema.Provider
)

const DebianTemplateUUID = "01000000-0000-4000-8000-000020070100"

func init() {
	TestAccProvider = Provider()
	TestAccProviderFactories = make(map[string]func() (tfprotov6.ProviderServer, error))
	TestAccProviderFactories["upcloud"] = func() (tfprotov6.ProviderServer, error) {
		factory, err := NewProviderServerFactory()
		return factory(), err
	}
}

func TestAccPreCheck(t *testing.T) {
	err := TestAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func IgnoreWhitespaceDiff(str string) *regexp.Regexp {
	ws := regexp.MustCompile(`\s+`)
	re := ws.ReplaceAllString(str, `\s+`)
	return regexp.MustCompile(re)
}

func CheckStringDoesNotChange(name, key string, expected *string) resource.TestCheckFunc {
	return func(s *tftest.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		actual := rs.Primary.Attributes[key]
		if *expected == "" {
			*expected = actual
		} else if actual != *expected {
			return fmt.Errorf(`expected %s to match previous value "%s", got "%s"`, key, *expected, actual)
		}
		return nil
	}
}

func UsingOpenTofu() bool {
	return strings.HasSuffix(os.Getenv("TF_ACC_TERRAFORM_PATH"), "tofu")
}
