package server

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func buildLoginOpts(ctx context.Context, data loginModel) (*request.LoginUser, string, diag.Diagnostics) {
	r := &request.LoginUser{}

	r.Username = data.User.ValueString()

	if !data.CreatePassword.IsNull() {
		if data.CreatePassword.ValueBool() {
			r.CreatePassword = "yes"
		} else {
			r.CreatePassword = "no"
		}
	}

	var keys []string
	diags := data.Keys.ElementsAs(ctx, &keys, false)
	for _, key := range keys {
		r.SSHKeys = append(r.SSHKeys, key)
	}

	return r, data.PasswordDelivery.ValueString(), diags
}

func buildSimpleBackupOpts(data *simpleBackupModel) string {
	if data == nil {
		return "no"
	}

	time := data.Time.ValueString()
	plan := data.Plan.ValueString()

	if time != "" && plan != "" {
		return fmt.Sprintf("%s,%s", time, plan)
	}

	return "no"
}

func buildStorageDeviceAddress(address, position string) string {
	if position != "" {
		return fmt.Sprintf("%s:%s", address, position)
	}

	return address
}

func buildNetworkOpts(ctx context.Context, data serverModel) (req []request.CreateServerInterface, diags diag.Diagnostics) {
	var ifaces []networkInterfaceModel
	diags.Append(data.NetworkInterfaces.ElementsAs(context.Background(), &ifaces, false)...)

	for _, iface := range ifaces {
		r := request.CreateServerInterface{
			Bootable: upcloud.FromBool(iface.Bootable.ValueBool()),
			Index:    int(iface.Index.ValueInt64()),
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  iface.IPAddressFamily.ValueString(),
					Address: iface.IPAddress.ValueString(),
				},
			},
			Network:           iface.Network.ValueString(),
			SourceIPFiltering: upcloud.FromBool(iface.SourceIPFiltering.ValueBool()),
			Type:              iface.Type.ValueString(),
		}

		if !iface.AdditionalIPAddresses.IsNull() {
			if r.Type != upcloud.NetworkTypePrivate {
				diags.AddError("Invalid configuration", "additional_ip_address can only be set for private network interfaces")
				return nil, diags
			}

			var additionalIPAddresses []additionalIPAddressModel
			diags.Append(iface.AdditionalIPAddresses.ElementsAs(ctx, &additionalIPAddresses, false)...)

			for _, ipAddress := range additionalIPAddresses {
				r.IPAddresses = append(r.IPAddresses, request.CreateServerIPAddress{
					Family:  ipAddress.IPAddressFamily.ValueString(),
					Address: ipAddress.IPAddress.ValueString(),
				})
			}
		}

		req = append(req, r)
	}

	return req, diags
}
