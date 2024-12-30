package server

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func findInterface(ifaces []upcloud.ServerInterface, index int) *upcloud.ServerInterface {
	for _, iface := range ifaces {
		if iface.Index == index {
			return &iface
		}
	}
	return nil
}

func findIPAddress(iface upcloud.ServerInterface, address string) *upcloud.IPAddress {
	for _, ip := range iface.IPAddresses {
		if ip.Address == address {
			return &ip
		}
	}
	return nil
}

func ipAndTypeMatches(api upcloud.ServerInterface, data networkInterfaceModel) bool {
	if api.Type != data.Type.ValueString() {
		return false
	}
	for _, ip := range api.IPAddresses {
		if ip.Address == data.IPAddress.ValueString() {
			return true
		}
	}
	return false
}

type ifaceChange struct {
	api       *upcloud.ServerInterface
	plan      *networkInterfaceModel
	planIndex *int
	state     *networkInterfaceModel
}

func intPtr(i int) *int {
	return &i
}

func matchInterfaces(api []upcloud.ServerInterface, state, plan []networkInterfaceModel) map[int]ifaceChange {
	m := make(map[int]ifaceChange)

	// For tracking interfaces that were matched by index.
	matchedPlanIfaces := make(map[int]bool)
	unmatchedIfaces := make(map[int]upcloud.ServerInterface)

	// Match interfaces by index.
	for i, apiIface := range api {
		change := ifaceChange{api: &apiIface}

		for j, planIface := range plan {
			if planIface.Index.ValueInt64() == int64(apiIface.Index) {
				change.plan = &planIface
				change.planIndex = intPtr(j)
				matchedPlanIfaces[j] = true
				break
			}
		}

		// If no match, this might be overridden by IP based matching below.
		m[i] = change

		if change.plan == nil {
			unmatchedIfaces[i] = apiIface
		}
	}

	// Match interfaces by IP address and type.
	for i, apiIface := range unmatchedIfaces {
		change := ifaceChange{api: &apiIface}

		for _, stateIface := range state {
			if ipAndTypeMatches(apiIface, stateIface) {
				change.state = &stateIface
				for k, planIface := range plan {
					if matchedPlanIfaces[k] {
						continue
					}
					if canModifyInterface(&stateIface, &planIface, &apiIface) {
						change.plan = &planIface
						change.planIndex = intPtr(k)
						matchedPlanIfaces[k] = true
						break
					}
				}
			}
		}
		m[i] = change
	}

	return m
}

func matchInterfacesToPlan(api []upcloud.ServerInterface, state, plan []networkInterfaceModel) map[int]ifaceChange {
	a := matchInterfaces(api, state, plan)
	b := make(map[int]ifaceChange)
	for _, change := range a {
		if change.planIndex != nil && change.plan != nil {
			b[*change.planIndex] = change
		}
	}
	return b
}

func canModifyInterface(state, plan *networkInterfaceModel, prev *upcloud.ServerInterface) bool {
	if plan == nil {
		return false
	}

	if prev.Type != plan.Type.ValueString() {
		return false
	}

	family := plan.IPAddressFamily
	if family.IsUnknown() && state != nil {
		family = state.IPAddressFamily
	}
	if len(prev.IPAddresses) > 0 && !family.IsUnknown() && prev.IPAddresses[0].Family != family.ValueString() {
		return false
	}
	if prev.Type == upcloud.NetworkTypePrivate {
		if prev.Network != plan.Network.ValueString() {
			return false
		}
		if !plan.IPAddress.IsUnknown() && !ipAndTypeMatches(*prev, *plan) {
			return false
		}
	}
	return true
}

func shouldModifyInterface(plan networkInterfaceModel, addresses request.CreateNetworkInterfaceIPAddressSlice, iface *upcloud.ServerInterface) bool {
	if iface.Index != int(plan.Index.ValueInt64()) {
		return true
	}

	if iface.Bootable.Bool() != plan.Bootable.ValueBool() {
		return true
	}

	if iface.SourceIPFiltering.Bool() != plan.SourceIPFiltering.ValueBool() {
		return true
	}

	for i, ip := range iface.IPAddresses {
		if i >= len(addresses) {
			return true
		}
		if ip.Family != addresses[i].Family {
			return true
		}
		// Additional IP addresses are only set for private networks
		if iface.Type != upcloud.NetworkTypePrivate {
			break
		}
		if ip.Address != addresses[i].Address {
			return true
		}
	}

	return false
}

func setInterfaceValues(ctx context.Context, iface *upcloud.Interface, ipInState types.String) (ni networkInterfaceModel, diags diag.Diagnostics) {
	additionalIPAddresses := []additionalIPAddressModel{}

	// IP addresses are not returned in deterministic order. If any of the IP addresses of the interface match the IP address in state, use that.
	if !ipInState.IsNull() {
		ip := ipInState.ValueString()
		for _, ipAddress := range iface.IPAddresses {
			if ipAddress.Address == ip {
				ni.IPAddressFamily = types.StringValue(ipAddress.Family)
				ni.IPAddress = types.StringValue(ipAddress.Address)
				ni.IPAddressFloating = types.BoolValue(ipAddress.Floating.Bool())
			}
		}
	}

	for i, ipAddress := range iface.IPAddresses {
		noKnownIP := ni.IPAddress.IsNull() || ni.IPAddress.IsUnknown()
		if i == 0 && noKnownIP {
			ni.IPAddressFamily = types.StringValue(ipAddress.Family)
			ni.IPAddress = types.StringValue(ipAddress.Address)
			ni.IPAddressFloating = types.BoolValue(ipAddress.Floating.Bool())
		} else if iface.Type == upcloud.NetworkTypePrivate && ipAddress.Address != ni.IPAddress.ValueString() {
			additionalIPAddresses = append(additionalIPAddresses, additionalIPAddressModel{
				IPAddressFamily:   types.StringValue(ipAddress.Family),
				IPAddress:         types.StringValue(ipAddress.Address),
				IPAddressFloating: types.BoolValue(ipAddress.Floating.Bool()),
			})
		}
	}
	ni.AdditionalIPAddresses, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: additionalIPAddressModel{}.AttributeTypes(),
	}, additionalIPAddresses)

	ni.Index = types.Int64Value(int64(iface.Index))
	ni.MACAddress = types.StringValue(iface.MAC)
	ni.Network = types.StringValue(iface.Network)
	ni.Type = types.StringValue(iface.Type)
	ni.Bootable = types.BoolValue(iface.Bootable.Bool())
	ni.SourceIPFiltering = types.BoolValue(iface.SourceIPFiltering.Bool())

	return
}

func updateServerNetworkInterfaces(ctx context.Context, svc *service.Service, state, plan *serverModel) (diags diag.Diagnostics) {
	uuid := plan.ID.ValueString()
	// Assert server is stopped
	s, err := svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: uuid,
	})
	if err != nil {
		diags.AddError("Unable to read server details", utils.ErrorDiagnosticDetail(err))
		return
	}
	if s.State != upcloud.ServerStateStopped {
		diags.AddError("Invalid server state", "Server needs to be stopped to alter networks")
		return
	}

	var networkInterfacesPlan []networkInterfaceModel
	diags.Append(plan.NetworkInterfaces.ElementsAs(ctx, &networkInterfacesPlan, false)...)

	var networkInterfacesState []networkInterfaceModel
	diags.Append(state.NetworkInterfaces.ElementsAs(ctx, &networkInterfacesState, false)...)

	indicesToKeep := map[int]bool{}
	n := len(networkInterfacesPlan)

	newNetworkInterfaces := make([]networkInterfaceModel, n)
	networkInterfaceChanges := matchInterfaces(s.Networking.Interfaces, networkInterfacesState, networkInterfacesPlan)
	modifiedInterfaces := make(map[int]*upcloud.Interface)
	indices := make(map[int]bool)

	// Try to modify existing server interfaces
	for i, iface := range s.Networking.Interfaces {
		indices[iface.Index] = true

		change := networkInterfaceChanges[i]
		stateVal := change.state
		planVal := change.plan

		// Remove interface if it has been removed from configuration
		if planVal == nil {
			err = svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
				ServerUUID: uuid,
				Index:      iface.Index,
			})
			if err != nil {
				diags.AddError("Unable to delete network interface", utils.ErrorDiagnosticDetail(err))
				return
			}
		}

		if !canModifyInterface(stateVal, planVal, &iface) {
			continue
		}

		addresses, d := addressesFromResourceData(ctx, svc, *planVal)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		modified := (*upcloud.Interface)(&iface)
		if shouldModifyInterface(*planVal, addresses, &iface) {
			req := request.ModifyNetworkInterfaceRequest{
				ServerUUID:   uuid,
				CurrentIndex: iface.Index,

				NewIndex:          int(planVal.Index.ValueInt64()),
				SourceIPFiltering: upcloud.FromBool(planVal.SourceIPFiltering.ValueBool()),
				Bootable:          upcloud.FromBool(planVal.Bootable.ValueBool()),
			}
			if iface.Type == upcloud.NetworkTypePrivate {
				req.IPAddresses = addresses
			}
			modified, err = svc.ModifyNetworkInterface(ctx, &req)
			if err != nil {
				diags.AddError("Unable to modify network interface", utils.ErrorDiagnosticDetail(err))
				return
			}
		}
		modifiedInterfaces[modified.Index] = modified
	}

	// Replace interfaces that can not be modified
	var d diag.Diagnostics
	for i := 0; i < n; i++ {
		val := networkInterfacesPlan[i]
		index := int(val.Index.ValueInt64())
		indicesToKeep[index] = true

		if modified := modifiedInterfaces[index]; modified != nil {
			newNetworkInterfaces[i], d = setInterfaceValues(ctx, modified, val.IPAddress)
			diags.Append(d...)
			continue
		}

		addresses, d := addressesFromResourceData(ctx, svc, val)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		t := val.Type.ValueString()
		network := ""
		if t == upcloud.NetworkTypePrivate {
			network = val.Network.ValueString()
		}

		if exists := indices[index]; exists {
			err = svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
				ServerUUID: uuid,
				Index:      index,
			})
			if err != nil {
				diags.AddError("Unable to delete network interface", utils.ErrorDiagnosticDetail(err))
				return
			}
		}

		iface, err := svc.CreateNetworkInterface(ctx, &request.CreateNetworkInterfaceRequest{
			ServerUUID:        uuid,
			Index:             index,
			Type:              t,
			NetworkUUID:       network,
			IPAddresses:       addresses,
			SourceIPFiltering: upcloud.FromBool(val.SourceIPFiltering.ValueBool()),
			Bootable:          upcloud.FromBool(val.Bootable.ValueBool()),
		})
		if err != nil {
			diags.AddError("Unable to create network interface", utils.ErrorDiagnosticDetail(err))
			return
		}
		newNetworkInterfaces[i], d = setInterfaceValues(ctx, iface, val.IPAddress)
		diags.Append(d...)
	}

	plan.NetworkInterfaces, d = types.ListValueFrom(ctx, plan.NetworkInterfaces.ElementType(ctx), newNetworkInterfaces)
	diags.Append(d...)
	return
}

func addressesFromResourceData(ctx context.Context, svc *service.Service, data networkInterfaceModel) (addresses request.CreateNetworkInterfaceIPAddressSlice, diags diag.Diagnostics) {
	ip := request.CreateNetworkInterfaceIPAddress{}
	if v := data.IPAddressFamily.ValueString(); v != "" {
		ip.Family = v
	}

	isPrivate := data.Type.ValueString() == upcloud.NetworkTypePrivate
	if isPrivate {
		net := data.Network.ValueString()
		if v := data.IPAddress.ValueString(); v != "" {
			ip.Address = v
			// Check if network contains IP or leave IP empty if network has DHCP is enabled.
			network, err := svc.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{UUID: net})
			if err != nil {
				diags.AddError("Unable to read network details", utils.ErrorDiagnosticDetail(err))
				return
			}
			ip.Address, err = resolveInterfaceIPAddress(network, v)
			if err != nil {
				diags.AddError("Unable to resolve interface IP address", err.Error())
				return
			}
		}
	}
	addresses = append(addresses, ip)

	if !data.AdditionalIPAddresses.IsNull() {
		var additionalIPAddresses []additionalIPAddressModel
		diags.Append(data.AdditionalIPAddresses.ElementsAs(ctx, &additionalIPAddresses, false)...)

		if len(additionalIPAddresses) > 0 && !isPrivate {
			diags.AddError("Invalid configuration", "additional_ip_address can only be set for private network interfaces")
			return
		}

		for _, ipAddress := range additionalIPAddresses {
			addresses = append(addresses, request.CreateNetworkInterfaceIPAddress{
				Family:  ipAddress.IPAddressFamily.ValueString(),
				Address: ipAddress.IPAddress.ValueString(),
			})
		}
	}
	return addresses, nil
}

func validateNetworkInterfaces(ctx context.Context, svc *service.Service, data serverModel) (diags diag.Diagnostics) {
	var ifaces []networkInterfaceModel
	diags.Append(data.NetworkInterfaces.ElementsAs(ctx, &ifaces, false)...)
	if diags.HasError() {
		return diags
	}

	for _, iface := range ifaces {
		_, d := addressesFromResourceData(ctx, svc, iface)
		diags.Append(d...)
	}
	return nil
}

func resolveInterfaceIPAddress(network *upcloud.Network, ipAddress string) (string, error) {
	ip, err := netip.ParseAddr(ipAddress)
	if err != nil {
		return ipAddress, err
	}

	var dhcpEnabled bool
	for _, n := range network.IPNetworks {
		ipNet, err := netip.ParsePrefix(n.Address)
		if err != nil {
			return ipAddress, err
		}
		if ipNet.Contains(ip) {
			return ipAddress, nil
		}
		if n.DHCP.Bool() {
			dhcpEnabled = true
		}
	}
	// We didn't find suitable network for IP address but there was DHCP service enabled which we can use.
	if dhcpEnabled {
		return "", nil
	}
	return "", fmt.Errorf("IP address %s is not valid for network %s (%s) which doesn't have DHCP enabled", ipAddress, network.Name, network.UUID)
}
