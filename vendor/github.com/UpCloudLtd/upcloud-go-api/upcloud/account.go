package upcloud

import "encoding/json"

// Account represents an account
type Account struct {
	Credits        float64        `json:"credits"`
	UserName       string         `json:"username"`
	ResourceLimits ResourceLimits `json:"resource_limits"`
}

// ResourceLimits represents an account's resource limits
type ResourceLimits struct {
	Cores               int `json:"cores"`
	DetachedFloatingIps int `json:"detached_floating_ips"`
	Memory              int `json:"memory"`
	Networks            int `json:"networks"`
	PublicIPv4          int `json:"public_ipv4"`
	PublicIPv6          int `json:"public_ipv6"`
	StorageHDD          int `json:"storage_hdd"`
	StorageSSD          int `json:"storage_ssd"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Account) UnmarshalJSON(b []byte) error {
	type localAccount Account

	v := struct {
		Account localAccount `json:"account"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Account(v.Account)

	return nil
}
