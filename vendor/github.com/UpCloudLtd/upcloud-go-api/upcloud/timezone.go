package upcloud

import "encoding/json"

// TimeZones represents a list of timezones
type TimeZones struct {
	TimeZones []string `json:"timezone"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *TimeZones) UnmarshalJSON(b []byte) error {
	type timezoneWrapper struct {
		TimeZones []string `json:"timezone"`
	}

	v := struct {
		TimeZones timezoneWrapper `json:"timezones"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	s.TimeZones = v.TimeZones.TimeZones

	return nil
}
