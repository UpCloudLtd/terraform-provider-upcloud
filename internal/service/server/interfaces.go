package server

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func findInterface(ifaces []upcloud.ServerInterface, index int) *upcloud.ServerInterface {
	for _, iface := range ifaces {
		if iface.Index == index {
			return &iface
		}
	}
	return nil
}

func interfacesToMap(ifaces []interface{}) map[int]interface{} {
	m := make(map[int]interface{})
	for i, iface := range ifaces {
		m[i] = iface
	}
	return m
}

func findInterfaceFromState(ifaces map[int]interface{}, index int, ip, t string) (int, map[string]interface{}, map[int]interface{}) {
	for i, iface := range ifaces {
		val := iface.(map[string]interface{})
		if val["index"].(int) == index {
			delete(ifaces, i)
			return i, iface.(map[string]interface{}), ifaces
		}
	}
	for i, iface := range ifaces {
		val := iface.(map[string]interface{})
		if val["ip_address"].(string) == ip && val["type"].(string) == t {
			delete(ifaces, i)
			return i, iface.(map[string]interface{}), ifaces
		}
	}
	return -1, nil, ifaces
}

func canModifyInterface(plan map[string]interface{}, prev *upcloud.ServerInterface) bool {
	if v, ok := plan["type"].(string); !ok || prev.Type != v {
		return false
	}
	if v, ok := plan["ip_address_family"].(string); !ok || len(prev.IPAddresses) > 0 && prev.IPAddresses[0].Family != v {
		return false
	}
	if prev.Type == upcloud.NetworkTypePrivate {
		if v, ok := plan["network"].(string); !ok || prev.Network != v {
			return false
		}
		if v, ok := plan["ip_address"].(string); !ok || len(prev.IPAddresses) > 0 && prev.IPAddresses[0].Address != v {
			return false
		}
	}
	return true
}

func shouldModifyInterface(plan map[string]interface{}, addresses request.CreateNetworkInterfaceIPAddressSlice, iface *upcloud.ServerInterface) bool {
	if iface.Index != plan["index"].(int) {
		return true
	}

	if iface.Bootable.Bool() != plan["bootable"].(bool) {
		return true
	}

	if iface.SourceIPFiltering.Bool() != plan["source_ip_filtering"].(bool) {
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

func setInterfaceValues(iface *upcloud.Interface, ipInState interface{}) map[string]interface{} {
	ni := make(map[string]interface{})
	additionalIPAddresses := []map[string]interface{}{}

	// IP addresses are not returned in deterministic order. If any of the IP addresses of the interface match the IP address in state, use that.
	if ip, ok := ipInState.(string); ok {
		for _, ipAddress := range iface.IPAddresses {
			if ipAddress.Address == ip {
				ni["ip_address_family"] = ipAddress.Family
				ni["ip_address"] = ipAddress.Address
				if !ipAddress.Floating.Empty() {
					ni["ip_address_floating"] = ipAddress.Floating.Bool()
				}
			}
		}
	}

	for i, ipAddress := range iface.IPAddresses {
		if i == 0 && ni["ip_address"] == nil {
			ni["ip_address_family"] = ipAddress.Family
			ni["ip_address"] = ipAddress.Address
			if !ipAddress.Floating.Empty() {
				ni["ip_address_floating"] = ipAddress.Floating.Bool()
			}
		} else if iface.Type == upcloud.NetworkTypePrivate && ipAddress.Address != ni["ip_address"].(string) {
			additionalIPAddress := map[string]interface{}{
				"ip_address":        ipAddress.Address,
				"ip_address_family": ipAddress.Family,
			}
			if !ipAddress.Floating.Empty() {
				additionalIPAddress["ip_address_floating"] = ipAddress.Floating.Bool()
			}
			additionalIPAddresses = append(additionalIPAddresses, additionalIPAddress)
		}
	}
	ni["additional_ip_address"] = additionalIPAddresses

	ni["index"] = iface.Index
	ni["mac_address"] = iface.MAC
	ni["network"] = iface.Network
	ni["type"] = iface.Type
	if !iface.Bootable.Empty() {
		ni["bootable"] = iface.Bootable.Bool()
	}
	if !iface.SourceIPFiltering.Empty() {
		ni["source_ip_filtering"] = iface.SourceIPFiltering.Bool()
	}

	return ni
}

func interfaceKey(i int) string {
	return fmt.Sprintf("network_interface.%d", i)
}

func updateServerNetworkInterfaces(ctx context.Context, svc *service.Service, d *schema.ResourceData) error {
	// Assert server is stopped
	s, err := svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return err
	}
	if s.State != upcloud.ServerStateStopped {
		return errors.New("server needs to be stopped to alter networks")
	}

	indicesToKeep := map[int]bool{}
	n, ok := d.Get("network_interface.#").(int)
	if !ok {
		return errors.New("unable to read network_interface count")
	}

	newNetworkInterfaces := make([]map[string]interface{}, n)
	networkInterfaces := interfacesToMap(d.Get("network_interface").([]interface{}))
	modifiedInterfaces := make(map[int]*upcloud.Interface)

	// Try to modify existing server interfaces
	var i int
	var val map[string]interface{}
	for _, iface := range s.Networking.Interfaces {
		i, val, networkInterfaces = findInterfaceFromState(networkInterfaces, iface.Index, iface.IPAddresses[0].Address, iface.Type)

		// Remove interface if it has been removed from configuration
		if val == nil {
			err = svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
				ServerUUID: d.Id(),
				Index:      iface.Index,
			})
			if err != nil {
				return err
			}
		}

		if !canModifyInterface(val, &iface) {
			continue
		}

		addresses, err := addressesFromResourceData(ctx, svc, d, interfaceKey(i))
		if err != nil {
			return err
		}

		modified := (*upcloud.Interface)(&iface)
		if shouldModifyInterface(val, addresses, &iface) {
			req := request.ModifyNetworkInterfaceRequest{
				ServerUUID:   d.Id(),
				CurrentIndex: iface.Index,

				NewIndex:          val["index"].(int),
				SourceIPFiltering: upcloud.FromBool(val["source_ip_filtering"].(bool)),
				Bootable:          upcloud.FromBool(val["bootable"].(bool)),
			}
			if iface.Type == upcloud.NetworkTypePrivate {
				req.IPAddresses = addresses
			}
			modified, err = svc.ModifyNetworkInterface(ctx, &req)
			if err != nil {
				return err
			}
		}
		modifiedInterfaces[modified.Index] = modified
	}

	// Replace interfaces that can not be modified
	for i := 0; i < n; i++ {
		key := interfaceKey(i)
		index := d.Get(key + ".index").(int)
		indicesToKeep[index] = true

		val, ok := d.Get(key).(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to read '%s' value", key)
		}

		if !d.HasChange(key) {
			newNetworkInterfaces[i] = val
			continue
		}

		if modified := modifiedInterfaces[index]; modified != nil {
			newNetworkInterfaces[i] = setInterfaceValues(modified, val["ip_address"])
			continue
		}

		addresses, err := addressesFromResourceData(ctx, svc, d, key)
		if err != nil {
			return err
		}

		t := val["type"].(string)
		network := ""
		if t == upcloud.NetworkTypePrivate {
			network = val["network"].(string)
		}

		err = svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
			ServerUUID: d.Id(),
			Index:      index,
		})
		if err != nil {
			var ucProb *upcloud.Problem
			if errors.As(err, &ucProb) && ucProb.Type != upcloud.ErrCodeInterfaceNotFound {
				return err
			}
		}

		iface, err := svc.CreateNetworkInterface(ctx, &request.CreateNetworkInterfaceRequest{
			ServerUUID:        d.Id(),
			Index:             index,
			Type:              val["type"].(string),
			NetworkUUID:       network,
			IPAddresses:       addresses,
			SourceIPFiltering: upcloud.FromBool(val["source_ip_filtering"].(bool)),
			Bootable:          upcloud.FromBool(val["bootable"].(bool)),
		})
		if err != nil {
			return err
		}
		newNetworkInterfaces[i] = setInterfaceValues(iface, val["ip_address"])
	}

	if err := d.Set("network_interface", newNetworkInterfaces); err != nil {
		return err
	}
	return nil
}

func addressesFromResourceData(ctx context.Context, svc *service.Service, d *schema.ResourceData, key string) (request.CreateNetworkInterfaceIPAddressSlice, error) {
	addresses := make(request.CreateNetworkInterfaceIPAddressSlice, 0)
	val, ok := d.Get(key).(map[string]interface{})
	if !ok {
		return addresses, fmt.Errorf("unable to read '%s' value", key)
	}

	ip := request.CreateNetworkInterfaceIPAddress{}
	if v, ok := val["ip_address_family"].(string); ok && v != "" {
		ip.Family = v
	}

	if v, ok := val["type"].(string); ok && v == upcloud.NetworkTypePrivate {
		net := ""
		if v, ok := val["network"].(string); ok {
			net = v
		}
		if v, ok := val["ip_address"].(string); ok && v != "" {
			ip.Address = v
			// If network has changed but ip hasn't, check if network contains IP or leave IP empty if network has DHCP is enabled.
			if d.HasChange(key+".network") && !d.HasChange(key+".ip_address") {
				network, err := svc.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{UUID: net})
				if err != nil {
					return nil, err
				}
				ip.Address, err = resolveInterfaceIPAddress(network, v)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	addresses = append(addresses, ip)

	if v, ok := d.GetOk(key + ".additional_ip_address"); ok {
		additionalIPAddresses := v.(*schema.Set).List()
		if len(additionalIPAddresses) > 0 && val["type"].(string) != upcloud.NetworkTypePrivate {
			return nil, fmt.Errorf("additional_ip_address can only be set for private network interfaces")
		}

		for _, v := range additionalIPAddresses {
			ipAddress := v.(map[string]interface{})

			addresses = append(addresses, request.CreateNetworkInterfaceIPAddress{
				Family:  ipAddress["ip_address_family"].(string),
				Address: ipAddress["ip_address"].(string),
			})
		}
	}
	return addresses, nil
}

func validateNetworkInterfaces(ctx context.Context, svc *service.Service, d *schema.ResourceData) error {
	nInf, ok := d.Get("network_interface.#").(int)
	if !ok {
		return errors.New("unable to read network_interface count")
	}
	for i := 0; i < nInf; i++ {
		key := fmt.Sprintf("network_interface.%d", i)
		if _, err := addressesFromResourceData(ctx, svc, d, key); err != nil {
			return err
		}
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
