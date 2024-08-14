package utils

import (
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func FilterZoneIDs(vs []upcloud.Zone, f func(upcloud.Zone) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v.ID)
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

// ExpandStrings expands a terraform list to slice of str
func ExpandStrings(data interface{}) []string {
	strSlice := []string{}
	for _, s := range data.(*schema.Set).List() {
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

// StorageAddressPositionFormat takes the address in any format and extracts the bus
// position only
func StorageAddressPositionFormat(address string) string {
	if ret := strings.SplitN(address, ":", 2); len(ret) > 0 {
		return ret[1]
	}

	return ""
}

func EnvKeyExists(keyPrefix string) bool {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, keyPrefix) {
			return true
		}
	}
	return false
}

func JoinSchemas(dst map[string]*schema.Schema, s ...map[string]*schema.Schema) map[string]*schema.Schema {
	if len(s) == 0 {
		return dst
	}
	for i := range s {
		for key, value := range s[i] {
			dst[key] = value
		}
	}
	return dst
}
