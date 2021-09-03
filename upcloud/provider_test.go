package upcloud_test

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
)

func TestProvider(t *testing.T) {
	if err := upcloud.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
