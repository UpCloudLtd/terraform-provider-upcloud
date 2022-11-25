package server

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func reconfigureServerNetworkInterfaces(ctx context.Context, svc *service.ServiceContext, d *schema.ResourceData) error {
	// assert server is stopped
	s, err := svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return err
	}
	if s.State != upcloud.ServerStateStopped {
		return errors.New("server needs to be stopped to alter networks")
	}

	reqs, err := networkInterfacesFromResourceData(ctx, svc, d)
	if err != nil {
		return err
	}

	// Try to preserve public (IPv4 or IPv6) and utility network interfaces so that IPs doesn't change
	preserveInterfaces := make(map[int]bool, 0)
	// flush interfaces
	for i, n := range s.Networking.Interfaces {
		if (n.Type == upcloud.NetworkTypePublic || n.Type == upcloud.NetworkTypeUtility) && len(reqs) > i && interfacesEquals(n, reqs[i]) {
			preserveInterfaces[n.Index] = true
			continue
		}
		if err := svc.DeleteNetworkInterface(ctx, &request.DeleteNetworkInterfaceRequest{
			ServerUUID: d.Id(),
			Index:      n.Index,
		}); err != nil {
			return fmt.Errorf("unable to delete interface #%d; %w", n.Index, err)
		}
	}
	// apply interfaces from state
	for _, r := range reqs {
		if _, ok := preserveInterfaces[r.Index]; ok && (r.Type == upcloud.NetworkTypePublic || r.Type == upcloud.NetworkTypeUtility) {
			continue
		}
		if _, err := svc.CreateNetworkInterface(ctx, &r); err != nil {
			return fmt.Errorf("unable to create interface #%d; %w", r.Index, err)
		}
	}
	return nil
}

func networkInterfacesFromResourceData(ctx context.Context, svc *service.ServiceContext, d *schema.ResourceData) ([]request.CreateNetworkInterfaceRequest, error) {
	rs := make([]request.CreateNetworkInterfaceRequest, 0)
	nInf, ok := d.Get("network_interface.#").(int)
	if !ok {
		return rs, errors.New("unable read network_interface count")
	}
	for i := 0; i < nInf; i++ {
		key := fmt.Sprintf("network_interface.%d", i)
		val, ok := d.Get(key).(map[string]interface{})
		if !ok {
			return rs, fmt.Errorf("unable to read '%s' value", key)
		}
		r := request.CreateNetworkInterfaceRequest{
			ServerUUID:  d.Id(),
			Index:       i + 1,
			IPAddresses: make(request.CreateNetworkInterfaceIPAddressSlice, 0),
		}
		if v, ok := val["type"].(string); ok {
			r.Type = v
		}
		ip := request.CreateNetworkInterfaceIPAddress{}
		if v, ok := val["ip_address_family"].(string); ok && v != "" {
			ip.Family = v
		}
		if r.Type == upcloud.NetworkTypePrivate {
			if v, ok := val["network"].(string); ok && v != "" {
				r.NetworkUUID = v
			}
			if v, ok := val["ip_address"].(string); ok && v != "" {
				ip.Address = v
				// If network has changed but ip hasn't, check if network contains IP or leave IP empty if network has DHCP is enabled.
				if d.HasChange(key+".network") && !d.HasChange(key+".ip_address") {
					network, err := svc.GetNetworkDetails(ctx, &request.GetNetworkDetailsRequest{UUID: r.NetworkUUID})
					if err != nil {
						return rs, err
					}
					ip.Address, err = resolveInterfaceIPAddress(network, v)
					if err != nil {
						return rs, err
					}
				}
			}
			if v, ok := val["source_ip_filtering"].(bool); ok {
				r.SourceIPFiltering = upcloud.FromBool(v)
			}
			if v, ok := val["bootable"].(bool); ok {
				r.Bootable = upcloud.FromBool(v)
			}
		}
		r.IPAddresses = append(r.IPAddresses, ip)
		rs = append(rs, r)
	}
	return rs, nil
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

func interfacesEquals(a upcloud.ServerInterface, b request.CreateNetworkInterfaceRequest) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Index != b.Index {
		return false
	}
	if len(a.IPAddresses) != 1 || len(b.IPAddresses) != 1 {
		return false
	}
	if a.IPAddresses[0].Family != b.IPAddresses[0].Family {
		return false
	}
	return true
}
