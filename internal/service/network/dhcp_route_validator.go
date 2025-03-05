package network

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type dhcpRouteValidator struct{}

func (v dhcpRouteValidator) Description(_ context.Context) string {
	return "must be valid CIDR with an optional -nexthop= definition"
}

func (v dhcpRouteValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dhcpRouteValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	re := regexp.MustCompile(`^([0-9.:/]+)(?:-nexthop=([0-9.:]+)){0,1}$`)
	matches := re.FindStringSubmatch(req.ConfigValue.ValueString())

	if len(matches) < 2 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid DHCP route", "DHCP route must be valid CIDR with an optional -nexthop= definition")
		return
	}

	cidr := matches[1]
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid CIDR in DHCP route", fmt.Sprintf("DHCP route must be valid CIDR (%s)", err.Error()))
	}

	if len(matches) == 3 && matches[2] != "" {
		nexthop := matches[2]
		if net.ParseIP(nexthop) == nil {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid nexthop IP in DHCP route", fmt.Sprintf("DHCP route nexthop must be a valid IP address, got %s", nexthop))
		}
	}
}
