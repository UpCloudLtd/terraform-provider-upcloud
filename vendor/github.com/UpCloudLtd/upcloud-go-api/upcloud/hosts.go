package upcloud

import (
	"encoding/json"
	"time"
)

// Hosts represents a GetHosts response
type Hosts struct {
	Hosts []Host `json:"hosts"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (n *Hosts) UnmarshalJSON(b []byte) error {
	type localHost Host
	type hostWrapper struct {
		Hosts []localHost `json:"host"`
	}

	v := struct {
		Hosts hostWrapper `json:"hosts"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, ln := range v.Hosts.Hosts {
		n.Hosts = append(n.Hosts, Host(ln))
	}

	return nil
}

// StatSlice is a slice of Stat structs
// This exsits to support a custom unmarshaller
type StatSlice []Stat

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *StatSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Networks []Stat `json:"stat"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.Networks

	return nil
}

// Host represents an individual Host in a response
type Host struct {
	ID             int       `json:"id"`
	Description    string    `json:"description"`
	Zone           string    `json:"zone"`
	WindowsEnabled Boolean   `json:"windows_enabled"`
	Stats          StatSlice `json:"stats"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Host) UnmarshalJSON(b []byte) error {
	type localHost Host

	v := struct {
		Host localHost `json:"host"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Host(v.Host)

	return nil
}

// Stat represents Host stats in a response
type Stat struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
