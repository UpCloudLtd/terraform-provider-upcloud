package upcloud

import "encoding/json"

// PriceZones represents a /price response
type PriceZones struct {
	PriceZones []PriceZone `json:"prices"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *PriceZones) UnmarshalJSON(b []byte) error {
	type serverWrapper struct {
		PriceZones []PriceZone `json:"zone"`
	}

	v := struct {
		PriceZones serverWrapper `json:"prices"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.PriceZones = v.PriceZones.PriceZones

	return nil
}

// PriceZone represents a price zone. A prize zone consists of multiple items that each have a price.
type PriceZone struct {
	Name string `json:"name"`

	Firewall               *Price `json:"firewall"`
	IORequestBackup        *Price `json:"io_request_backup"`
	IORequestMaxIOPS       *Price `json:"io_request_maxiops"`
	IPv4Address            *Price `json:"ipv4_address"`
	IPv6Address            *Price `json:"ipv6_address"`
	PublicIPv4BandwidthIn  *Price `json:"public_ipv4_bandwidth_in"`
	PublicIPv4BandwidthOut *Price `json:"public_ipv4_bandwidth_out"`
	PublicIPv6BandwidthIn  *Price `json:"public_ipv6_bandwidth_in"`
	PublicIPv6BandwidthOut *Price `json:"public_ipv6_bandwidth_out"`
	ServerCore             *Price `json:"server_core"`
	ServerMemory           *Price `json:"server_memory"`
	ServerPlan1xCPU1GB     *Price `json:"server_plan_1xCPU-1GB"`
	ServerPlan2xCPU2GB     *Price `json:"server_plan_1xCPU-2GB"`
	ServerPlan4xCPU4GB     *Price `json:"server_plan_4xCPU-4GB"`
	ServerPlan6xCPU8GB     *Price `json:"server_plan_6xCPU-8GB"`
	StorageBackup          *Price `json:"storage_backup"`
	StorageMaxIOPS         *Price `json:"storage_maxiops"`
	StorageTemplate        *Price `json:"storage_template"`
}

// Price represents a price
type Price struct {
	Amount int     `json:"amount"`
	Price  float64 `json:"price"`
}
