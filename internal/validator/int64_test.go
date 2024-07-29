package validator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"

	fwvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func TestDivisibleBy(t *testing.T) {
	divisor := int64(1024)
	validValues := []int64{
		1024,
		2048,
		4096,
		10240,
		131072,
	}
	invalidValues := []int64{
		1025,
		1000,
		2000,
		10000,
		13000,
	}

	v := DivisibleBy(divisor)

	for _, value := range validValues {
		req := fwvalidator.Int64Request{
			Path:        path.Empty(),
			ConfigValue: types.Int64Value(value),
		}
		resp := fwvalidator.Int64Response{}

		v.ValidateInt64(context.Background(), req, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("DivisibleBy failed with valid value'%d: %s", value, resp.Diagnostics.Errors())
		}
	}

	for _, value := range invalidValues {
		req := fwvalidator.Int64Request{
			Path:        path.Empty(),
			ConfigValue: types.Int64Value(value),
		}
		resp := fwvalidator.Int64Response{}

		v.ValidateInt64(context.Background(), req, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("DivisibleBy did not fail with invalid value %d", value)
		}
	}
}
