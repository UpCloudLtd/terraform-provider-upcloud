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

func findInterface(ifaces []upcloud.ServerInterface, index int, ip string) *upcloud.ServerInterface {
	for _, iface := range ifaces {
		if iface.Index == index {
			return &iface
		}
	}
	for _, iface := range ifaces {
		if len(iface.IPAddresses) > 0 && iface.IPAddresses[0].Address == ip {
			return &iface
		}
	}
	return nil
}

func canModifyInterface(plan map[string]interface{}, prev *upcloud.ServerInterface) bool {
	if prev.Type != plan["type"].(string) {
		return false
	}
	if prev.Network != plan["network"].(string) {
		return false
	}
	if len(prev.IPAddresses) > 0 && prev.IPAddresses[0].Family != plan["ip_address_family"].(string) {
		return false
	}
	return true
}

func setInterfaceValues(iface *upcloud.Interface) map[string]interface{} {
	ni := make(map[string]interface{})
	additionalIPAddresses := []map[string]interface{}{}

	for i, ipAddress := range iface.IPAddresses {
		if i == 0 {
			ni["ip_address_family"] = ipAddress.Family
			ni["ip_address"] = ipAddress.Address
			if !ipAddress.Floating.Empty() {
				ni["ip_address_floating"] = ipAddress.Floating.Bool()
			}
		} else if iface.Type == upcloud.NetworkTypePrivate {
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

	networkInterfaces := make([]map[string]interface{}, n)

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("network_interface.%d", i)
		index := d.Get(key + ".index").(int)
		indicesToKeep[index] = true

		val, ok := d.Get(key).(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to read '%s' value", key)
		}

		if !d.HasChange(key) {
			networkInterfaces[i] = val
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

		prev := findInterface(s.Networking.Interfaces, index, val["ip_address"].(string))
		var iface *upcloud.Interface
		if prev == nil || !canModifyInterface(val, prev) {
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

			iface, err = svc.CreateNetworkInterface(ctx, &request.CreateNetworkInterfaceRequest{
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
		} else {
			iface, err = svc.ModifyNetworkInterface(ctx, &request.ModifyNetworkInterfaceRequest{
				ServerUUID:   d.Id(),
				CurrentIndex: prev.Index,

				NewIndex:          val["index"].(int),
				IPAddresses:       addresses,
				SourceIPFiltering: upcloud.FromBool(val["source_ip_filtering"].(bool)),
				Bootable:          upcloud.FromBool(val["bootable"].(bool)),
			})
			if err != nil {
				return err
			}
		}
		networkInterfaces[i] = setInterfaceValues(iface)
	}

	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return err
	}

	// Remove interfaces that are removed from configuration
	for _, iface := range s.Networking.Interfaces {
		if !indicesToKeep[iface.Index] {
			err = svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
				ServerUUID: d.Id(),
				Index:      iface.Index,
			})
			if err != nil {
				return err
			}
		}
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
