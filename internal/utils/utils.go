package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

// ExpandStrings expands a terraform list to slice of str
func ExpandStrings(data interface{}) []string {
	strSlice := []string{}
	for _, s := range data.([]interface{}) {
		strSlice = append(strSlice, s.(string))
	}

	return strSlice
}

// SetOfStringsToSlice transforms a terraform set of strings to a slice of strings
func SetOfStringsToSlice(ctx context.Context, data interface{}) ([]string, error) {
	result := []string{}
	providerErrMsg := "provider error: failed to transform set data"
	debugLogPrefix := "transforming set of strings into slice failed;"

	stringsSet, ok := data.(*schema.Set)
	if !ok {
		tflog.Debug(ctx, fmt.Sprintf("%s expected input data to be a schema.TypeSet but received %T", debugLogPrefix, data))
		return result, fmt.Errorf(providerErrMsg)
	}

	for _, val := range stringsSet.List() {
		valStr, ok := val.(string)
		if !ok {
			tflog.Debug(ctx, fmt.Sprintf("%s expected set elements to be of type string but received %T", debugLogPrefix, val))
			return result, fmt.Errorf(providerErrMsg)
		}

		result = append(result, valStr)
	}

	return result, nil
}

// MapOfStringsToLabelSlice transforms a terraform map of strings to a LabelSlice
func MapOfStringsToLabelSlice(ctx context.Context, data interface{}) (upcloud.LabelSlice, error) {
	result := upcloud.LabelSlice{}
	providerErrMsg := "provider error: failed to transform labels data"
	debugLogPrefix := "transforming map of strings into labels slice failed;"

	labelsMap, ok := data.(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, fmt.Sprintf("%s expected input data to be a map of strings but received %T", debugLogPrefix, data))
		return result, fmt.Errorf(providerErrMsg)
	}

	for k, v := range labelsMap {
		value, ok := v.(string)
		if !ok {
			tflog.Debug(ctx, fmt.Sprintf("%s expected map elements to be of type string but received %T", debugLogPrefix, v))
			return result, fmt.Errorf(providerErrMsg)
		}

		result = append(result, upcloud.Label{
			Key:   k,
			Value: value,
		})
	}

	return result, nil
}

// LabelSliceToMap transorms `upcloud.LabelSlice` into a map of strings.
// This can be used to set labels fetched from the API into a state
func LabelSliceToMap(data upcloud.LabelSlice) map[string]string {
	result := map[string]string{}
	for _, label := range data {
		result[label.Key] = label.Value
	}

	return result
}

// SliceOfStringToServerUUIDSlice converts slice of strings into `upcloud.ServerUUIDSlice`
func SliceOfStringToServerUUIDSlice(strs []string) upcloud.ServerUUIDSlice {
	result := make(upcloud.ServerUUIDSlice, len(strs))
	copy(result, strs)
	return result
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
