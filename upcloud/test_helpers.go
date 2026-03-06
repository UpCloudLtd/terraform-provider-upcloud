package upcloud

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tftest "github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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

func CaptureServerStartTime(resourceName, keyDir string, captured *string) resource.TestCheckFunc {
	return func(s *tftest.State) error {
		startTime, err := readServerStartTime(s, resourceName, keyDir)
		if err != nil {
			return err
		}

		*captured = startTime

		return nil
	}
}

func CheckServerStartTime(resourceName, keyDir string, captured *string, operation string, expectChanged bool) resource.TestCheckFunc {
	return func(s *tftest.State) error {
		if *captured == "" {
			return fmt.Errorf("captured start time for %s is empty", resourceName)
		}

		currentStartTime, err := readServerStartTime(s, resourceName, keyDir)
		if err != nil {
			return err
		}

		if expectChanged && currentStartTime == *captured {
			return fmt.Errorf("server was not restarted after %s: original=%s current=%s", operation, *captured, currentStartTime)
		}

		if !expectChanged && currentStartTime != *captured {
			return fmt.Errorf("server was restarted after %s: original=%s current=%s", operation, *captured, currentStartTime)
		}

		return nil
	}
}

func readServerStartTime(s *tftest.State, resourceName, keyDir string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("root module has no resource called %s", resourceName)
	}

	ipAddress := rs.Primary.Attributes["network_interface.0.ip_address"]
	if ipAddress == "" {
		return "", fmt.Errorf("resource %s has empty network_interface.0.ip_address", resourceName)
	}
	privateKey, err := os.ReadFile(filepath.Join(keyDir, "id_rsa"))
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	var lastErr error

	for range 30 {
		callback, callbackErr := hostKeyCallback(ipAddress, keyDir)
		if callbackErr != nil {
			lastErr = callbackErr
			time.Sleep(5 * time.Second)
			continue
		}

		config := &ssh.ClientConfig{
			User:            "root",
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: callback,
			Timeout:         10 * time.Second,
		}

		client, dialErr := ssh.Dial("tcp", fmt.Sprintf("%s:22", ipAddress), config)
		if dialErr != nil {
			lastErr = dialErr
			time.Sleep(5 * time.Second)
			continue
		}

		session, sessionErr := client.NewSession()
		if sessionErr != nil {
			_ = client.Close()
			lastErr = sessionErr
			time.Sleep(5 * time.Second)
			continue
		}

		output, cmdErr := session.Output("uptime -s")
		_ = session.Close()
		_ = client.Close()
		if cmdErr != nil {
			lastErr = cmdErr
			time.Sleep(5 * time.Second)
			continue
		}

		return strings.TrimSpace(string(output)), nil
	}

	return "", fmt.Errorf("failed to read server start time from %s: %w", ipAddress, lastErr)
}

func hostKeyCallback(ipAddress, keyDir string) (ssh.HostKeyCallback, error) {
	knownHostsPath := filepath.Join(keyDir, "known_hosts")

	cmd := exec.Command("ssh-keyscan", "-T", "5", ipAddress)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to scan SSH host key for %s: %w", ipAddress, err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, fmt.Errorf("ssh-keyscan returned empty host key for %s", ipAddress)
	}

	if writeErr := os.WriteFile(knownHostsPath, output, 0o600); writeErr != nil {
		return nil, fmt.Errorf("failed to write known_hosts file: %w", writeErr)
	}

	callback, callbackErr := knownhosts.New(knownHostsPath)
	if callbackErr != nil {
		return nil, fmt.Errorf("failed to create known_hosts callback: %w", callbackErr)
	}

	return callback, nil
}
