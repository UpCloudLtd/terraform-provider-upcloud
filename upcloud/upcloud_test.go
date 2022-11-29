package upcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/config"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/sandbox"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/service"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var provider *schema.Provider

var testAccProviderFactories = configureProviderFactories()

func configureProviderFactories() map[string]func() (*schema.Provider, error) {
	tagsProvider := Provider()
	return map[string]func() (*schema.Provider, error){
		"upcloud": func() (*schema.Provider, error) {
			return provider, nil
		},
		"upcloud-tags": func() (*schema.Provider, error) {
			return tagsProvider, nil
		},
	}
}

func TestMain(m *testing.M) {
	username := os.Getenv("UPCLOUD_USERNAME")
	password := os.Getenv("UPCLOUD_PASSWORD")
	svc, cancelSandbox, err := setupSandbox(username, password)
	if err != nil {
		log.Fatal(err)
	}
	provider = Provider()
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return svc, nil
	}
	var exitCode int
	defer func() {
		cancelSandbox()
		os.Exit(exitCode)
	}()
	// resource.TestMain(m) calls os.Exit which would make our defer function efectless
	// resource.TestMain(m)
	exitCode = m.Run()
}

func setupSandbox(username, password string) (svc *service.Service, cancel func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	r := retryablehttp.NewClient()
	sb := sandbox.NewWithHTTPClient(username, password, r.HTTPClient)
	user, err := sb.Create(ctx)
	if err != nil {
		return svc, cancel, err
	}
	user.UserAgent = fmt.Sprintf("terraform-provider-upcloud/%s-test", config.Version)
	return service.New(user), func() {
		if sb == nil {
			return
		}
		if err := sb.Delete(context.Background()); err != nil {
			log.Print(err)
		}
	}, nil
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_USERNAME"); v == "" {
		t.Fatal("UPCLOUD_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_PASSWORD must be set for acceptance tests")
	}
}
