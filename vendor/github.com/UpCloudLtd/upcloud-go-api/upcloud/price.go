package upcloud

// PrizeZones represents a /price response
type PrizeZones struct {
	PrizeZones []PrizeZone `xml:"zone"`
}

// PrizeZone represents a price zone. A prize zone consists of multiple items that each have a price.
type PrizeZone struct {
	Name string `xml:"name"`

	Firewall               *Price `xml:"firewall"`
	IORequestBackup        *Price `xml:"io_request_backup"`
	IORequestMaxIOPS       *Price `xml:"io_request_maxiops"`
	IPv4Address            *Price `xml:"ipv4_address"`
	IPv6Address            *Price `xml:"ipv6_address"`
	PublicIPv4BandwidthIn  *Price `xml:"public_ipv4_bandwidth_in"`
	PublicIPv4BandwidthOut *Price `xml:"public_ipv4_bandwidth_out"`
	PublicIPv6BandwidthIn  *Price `xml:"public_ipv6_bandwidth_in"`
	PublicIPv6BandwidthOut *Price `xml:"public_ipv6_bandwidth_out"`
	ServerCore             *Price `xml:"server_core"`
	ServerMemory           *Price `xml:"server_memory"`
	ServerPlan1xCPU1GB     *Price `xml:"server_plan_1xCPU-1GB"`
	ServerPlan2xCPU2GB     *Price `xml:"server_plan_2xCPU-2GB"`
	ServerPlan4xCPU4GB     *Price `xml:"server_plan_4xCPU-4GB"`
	ServerPlan6xCPU8GB     *Price `xml:"server_plan_6xCPU-8GB"`
	StorageBackup          *Price `xml:"storage_backup"`
	StorageMaxIOPS         *Price `xml:"storage_maxiops"`
	StorageTemplate        *Price `xml:"storage_template"`
}

// Price represents a price
type Price struct {
	Amount int     `xml:"amount"`
	Price  float64 `xml:"price"`
}
