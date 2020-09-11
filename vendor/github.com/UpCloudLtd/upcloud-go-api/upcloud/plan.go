package upcloud

import "encoding/json"

// Plans represents a /plan response
type Plans struct {
	Plans []Plan `json:"plans"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Plans) UnmarshalJSON(b []byte) error {
	type planWrapper struct {
		Plans []Plan `json:"plan"`
	}

	v := struct {
		Plans planWrapper `json:"plans"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.Plans = v.Plans.Plans

	return nil
}

// Plan represents a pre-configured server configuration plan
type Plan struct {
	CoreNumber       int    `json:"core_number"`
	MemoryAmount     int    `json:"memory_amount"`
	Name             string `json:"name"`
	PublicTrafficOut int    `json:"public_traffic_out"`
	StorageSize      int    `json:"storage_size"`
	StorageTier      string `json:"storage_tier"`
}
