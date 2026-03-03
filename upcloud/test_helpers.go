package upcloud

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tftest "github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/crypto/ssh"
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

func GenerateSSHKeyPair(keyDir string) error {
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyFile := filepath.Join(keyDir, "id_rsa")
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	if err := os.WriteFile(privateKeyFile, privateKeyPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pub)
	publicKeyFile := filepath.Join(keyDir, "id_rsa.pub")
	if err := os.WriteFile(publicKeyFile, publicKeyBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func UptimeProvisioner(keyDir string, captureUptime bool, checkUptime bool, operation string) string {
	if captureUptime {
		provisioner := `
			provisioner "remote-exec" {
				inline = [
					"uptime -s > /tmp/server_start_time.txt",
				]
				connection {
					type        = "ssh"
					user        = "root"
					host        = self.network_interface[0].ip_address
					private_key = file("%s/id_rsa")
				}
			}
		`
		return fmt.Sprintf(provisioner, keyDir)
	}

	if checkUptime {
		provisioner := `
			provisioner "remote-exec" {
				inline = [
					"if [ -f /tmp/server_start_time.txt ]; then",
					"  ORIGINAL_START_TIME=$(cat /tmp/server_start_time.txt)",
					"  CURRENT_START_TIME=$(uptime -s)",
					"  echo \"Original start time: $ORIGINAL_START_TIME\"",
					"  echo \"Current start time: $CURRENT_START_TIME\"",
					"  if [ \"$ORIGINAL_START_TIME\" = \"$CURRENT_START_TIME\" ]; then",
					"    echo 'SUCCESS: Server was not restarted after %s'",
					"    exit 0",
					"  else",
					"    echo 'ERROR: Server was restarted after %s'",
					"    exit 1",
					"  fi",
					"else",
					"  echo 'ERROR: Could not find server start time file'",
					"  exit 1",
					"fi",
				]
				connection {
					type        = "ssh"
					user        = "root"
					host        = self.network_interface[0].ip_address
					private_key = file("%s/id_rsa")
				}
			}
		`
		return fmt.Sprintf(provisioner, operation, operation, keyDir)
	}

	return ""
}
