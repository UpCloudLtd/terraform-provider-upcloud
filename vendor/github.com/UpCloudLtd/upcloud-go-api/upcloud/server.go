package upcloud

// Constants
const (
	ServerStateStarted     = "started"
	ServerStateStopped     = "stopped"
	ServerStateMaintenance = "maintenance"
	ServerStateError       = "error"

	VideoModelVGA    = "vga"
	VideoModelCirrus = "cirrus"

	StopTypeSoft = "soft"
	StopTypeHard = "hard"
)

// ServerConfigurations represents a /server_size response
type ServerConfigurations struct {
	ServerConfigurations []ServerConfiguration `xml:"server_size"`
}

// ServerConfiguration represents a server configuration
type ServerConfiguration struct {
	CoreNumber   int `xml:"core_number"`
	MemoryAmount int `xml:"memory_amount"`
}

// Servers represents a /server response
type Servers struct {
	Servers []Server `xml:"server"`
}

// Server represents a server
type Server struct {
	CoreNumber   int      `xml:"core_number"`
	Hostname     string   `xml:"hostname"`
	License      float64  `xml:"license"`
	MemoryAmount int      `xml:"memory_amount"`
	Plan         string   `xml:"plan"`
	Progress     int      `xml:"progress"`
	State        string   `xml:"state"`
	Tags         []string `xml:"tags>tag"`
	Title        string   `xml:"title"`
	UUID         string   `xml:"uuid"`
	Zone         string   `xml:"zone"`
}

// ServerDetails represents details about a server
type ServerDetails struct {
	Server

	BootOrder  string `xml:"boot_order"`
	CoreNumber int    `xml:"core_number"`
	// TODO: Convert to boolean
	Firewall       string                `xml:"firewall"`
	Host           int                   `xml:"host"`
	IPAddresses    []IPAddress           `xml:"ip_addresses>ip_address"`
	NICModel       string                `xml:"nic_model"`
	StorageDevices []ServerStorageDevice `xml:"storage_devices>storage_device"`
	Timezone       string                `xml:"timezone"`
	VideoModel     string                `xml:"video_model"`
	// TODO: Convert to boolean
	VNC         string `xml:"vnc"`
	VNCHost     string `xml:"vnc_host"`
	VNCPassword string `xml:"vnc_password"`
	VNCPort     int    `xml:"vnc_port"`
}
