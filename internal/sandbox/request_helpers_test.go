package sandbox

import (
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud/request"
)

func createRouterRequest() *request.CreateRouterRequest {
	return &request.CreateRouterRequest{
		Name: tempName("test"),
	}
}

func createNetworkRequest(zone, routerID string) *request.CreateNetworkRequest {
	return &request.CreateNetworkRequest{
		Name:   tempName("net-1"),
		Zone:   zone,
		Router: routerID,
		IPNetworks: []upcloud.IPNetwork{
			{
				Address: "10.0.100.0/24",
				Family:  upcloud.IPAddressFamilyIPv4,
				DHCP:    1,
			},
		},
	}
}

func createServerRequest(zone, IP, sdnID string) *request.CreateServerRequest {
	return &request.CreateServerRequest{
		Zone:             zone,
		Title:            tempName("test-server"),
		Hostname:         tempName("test"),
		PasswordDelivery: request.PasswordDeliveryNone,
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateServerStorageDeviceActionClone,
				Storage: "01000000-0000-4000-8000-000020060100",
				Title:   "disk1",
				Size:    10,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: []request.CreateServerInterface{
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family:  upcloud.IPAddressFamilyIPv4,
							Address: IP,
						},
					},
					Type:    upcloud.NetworkTypePrivate,
					Network: sdnID,
				},
			},
		},
	}
}

func createStorageRequest(zone string) *request.CreateStorageRequest {
	return &request.CreateStorageRequest{
		Size:  10,
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: tempName("test"),
		Zone:  zone,
	}
}

func createObjectStorageRequest(zone string) *request.CreateObjectStorageRequest {
	return &request.CreateObjectStorageRequest{
		Name:      tempName("test"),
		Zone:      zone,
		AccessKey: password(32),
		SecretKey: password(32),
		Size:      250,
	}
}

func createManagedDatabaseRequest(zone string) *request.CreateManagedDatabaseRequest {
	return &request.CreateManagedDatabaseRequest{
		HostNamePrefix: tempName("test"),
		Plan:           "1x1xCPU-2GB-25GB",
		Title:          "test-title",
		Type:           "pg",
		Zone:           zone,
	}
}

func createLoadBalancerRequest(zone, srvIP, sdnID string) *request.CreateLoadBalancerRequest {
	return &request.CreateLoadBalancerRequest{
		Name: tempName("test"),
		Plan: "development",
		Zone: zone,
		Networks: []request.LoadBalancerNetwork{
			{
				Name:   "pubtest",
				Type:   upcloud.LoadBalancerNetworkTypePublic,
				Family: "IPv4",
			},
			{
				Name:   "test",
				Type:   upcloud.LoadBalancerNetworkTypePrivate,
				Family: "IPv4",
				UUID:   sdnID,
			},
		},
		ConfiguredStatus: upcloud.LoadBalancerConfiguredStatusStarted,
		Frontends:        []request.LoadBalancerFrontend{},
		Backends: []request.LoadBalancerBackend{{
			Name:     "test",
			Resolver: "",
			Members: []request.LoadBalancerBackendMember{{
				Name:    "test",
				Enabled: true,
				Type:    upcloud.LoadBalancerBackendMemberTypeStatic,
				IP:      srvIP,
				Port:    80,
			}},
			Properties: &upcloud.LoadBalancerBackendProperties{},
		}},
		Resolvers: []request.LoadBalancerResolver{},
	}
}
