package utils

import (
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

func FilterZoneIds(vs []upcloud.Zone, f func(upcloud.Zone) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v.ID)
		}
	}
	return vsf
}

func FilterZones(vs []upcloud.Zone, f func(upcloud.Zone) bool) []upcloud.Zone {
	vsf := make([]upcloud.Zone, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterNetworks(vs []upcloud.Network, fns ...func(upcloud.Network) (bool, error)) ([]upcloud.Network, error) {
	vsf := []upcloud.Network{}

	for _, v := range vs {
		matched := true
		for _, fn := range fns {
			m, err := fn(v)
			if err != nil {
				return nil, err
			}

			if !m {
				matched = false
				break
			}
		}

		if matched {
			vsf = append(vsf, v)
		}
	}

	return vsf, nil
}

// WithRetry attempts to call the provided function until it has been successfully called or the number of calls exceeds retries delaying the consecutive calls by given delay
func WithRetry(fn func() (interface{}, error), retries int, delay time.Duration) (interface{}, error) {
	var err error
	var res interface{}
	for count := 0; true; count++ {
		if delay > 0 {
			time.Sleep(delay)
		}
		if count >= retries {
			break
		}
		res, err = fn()
		if err == nil {
			return res, nil
		}
		continue
	}
	return nil, err
}

// ExpandStrings expands a terraform interface to slice of str
func ExpandStrings(data interface{}) []string {
	strSlice := []string{}
	for _, s := range data.([]interface{}) {
		strSlice = append(strSlice, s.(string))
	}

	return strSlice
}

// StorageAddressFormat takes the address in any format and extracts the bus
// type only (ide/scsi/virtio)
func StorageAddressFormat(address string) string {
	if ret := strings.Split(address, ":"); len(ret) > 0 {
		return ret[0]
	}

	return ""
}
