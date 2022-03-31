package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func EnvKeyExists(keyPrefix string) bool {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, keyPrefix) {
			return true
		}
	}
	return false
}

func JoinSchemas(src, dst map[string]*schema.Schema) map[string]*schema.Schema {
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func ValidateDomainName(hostname string) error {
	const (
		minLen      int = 1
		maxLen      int = 253
		labelMaxLen int = 63
	)
	l := len(hostname)

	if l > maxLen || l < minLen {
		return fmt.Errorf("%s length %d is not in the range %d - %d", hostname, l, minLen, maxLen)
	}

	if hostname[0] == '.' || hostname[0] == '-' {
		return fmt.Errorf("%s starts with dot or hyphen", hostname)
	}

	if hostname[l-1] == '.' || hostname[l-1] == '-' {
		return fmt.Errorf("%s ends with dot or hyphen", hostname)
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen (either one is required)
	labelLen := 0

	for i := 0; i < l; i++ {
		c := hostname[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			labelLen++
		case '0' <= c && c <= '9':
			labelLen++
		case c == '-':
			if last == '.' {
				return fmt.Errorf("'%s' character before hyphen cannot be dot", hostname[0:i+1])
			}
			labelLen++
			nonNumeric = true
		case c == '.':
			if last == '.' || last == '-' {
				return fmt.Errorf("'%s' character before dot cannot be dot or hyphen", hostname[0:i+1])
			}
			if labelLen > labelMaxLen || labelLen == 0 {
				return fmt.Errorf("'%s' label is not in the range %d - %d", hostname[0:i+1], minLen, labelMaxLen)
			}
			labelLen = 0
		default:
			return fmt.Errorf("%s contains illegal characters", hostname)
		}
		last = c
	}

	if labelLen > labelMaxLen {
		return fmt.Errorf("%s label is not in the range %d - %d", hostname, minLen, labelMaxLen)
	}

	if !nonNumeric {
		return fmt.Errorf("%s contains only numeric labels", hostname)
	}

	return nil
}
