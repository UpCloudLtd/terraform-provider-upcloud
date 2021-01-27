package server

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildNetworkOpts(d *schema.ResourceData, meta interface{}) ([]request.CreateServerInterface, error) {
	ifaces := []request.CreateServerInterface{}

	niCount := d.Get("network_interface.#").(int)
	for i := 0; i < niCount; i++ {
		keyRoot := fmt.Sprintf("network_interface.%d.", i)

		iface := request.CreateServerInterface{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  d.Get(keyRoot + "ip_address_family").(string),
					Address: d.Get(keyRoot + "ip_address").(string),
				},
			},
			Type: d.Get(keyRoot + "type").(string),
		}

		iface.SourceIPFiltering = upcloud.FromBool(d.Get(keyRoot + "source_ip_filtering").(bool))
		iface.Bootable = upcloud.FromBool(d.Get(keyRoot + "bootable").(bool))

		if v, ok := d.GetOk(keyRoot + "network"); ok {
			iface.Network = v.(string)
		}

		ifaces = append(ifaces, iface)
	}

	return ifaces, nil
}
