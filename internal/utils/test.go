package utils

import (
	"os"
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/upcloud-go-api/credentials"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
)

// ReadTestDataFile reads testdata from file to a string. Fails tests with Fatal, if reading the file fails.
func ReadTestDataFile(t *testing.T, name string) string {
	testdata, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(testdata)
}

func NewServiceWithCredentialsFromEnv(t *testing.T) *service.Service {
	t.Helper()

	creds, err := credentials.Parse(credentials.Credentials{})
	if err != nil {
		t.Skip("UpCloud credentials not set.")
		return nil
	}

	cfg := config.NewFromCredentials(creds)
	authFn, err := cfg.WithAuth()
	if err != nil {
		t.Skip("UpCloud credentials not set.")
	}
	return service.New(client.New("", "", authFn))
}
