package utils

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// stateIllegalRetryInterval is a variable so that tests can shorten it.
var stateIllegalRetryInterval = 5 * time.Second

// RetryOnServerStateIllegal calls fn until it does not fail with SERVER_STATE_ILLEGAL error, ctx is done, or 10 minutes have passed.
// The API returns this error, for example, when assigning an IP address to a server that is in maintenance state.
func RetryOnServerStateIllegal[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	return retryOnErrorCode(ctx, upcloud.ErrCodeServerStateIllegal, fn)
}

// RetryOnStorageStateIllegal calls fn until it does not fail with STORAGE_STATE_ILLEGAL error, ctx is done, or 10 minutes have passed.
// The API returns this error, for example, when attaching a storage that another server is still detaching.
func RetryOnStorageStateIllegal[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	return retryOnErrorCode(ctx, upcloud.ErrCodeStorageStateIllegal, fn)
}

// RetryOnStorageAttached calls fn until it does not fail with STORAGE_ATTACHED error, ctx is done, or 2 minutes have passed.
// The API returns this error when attaching a storage that another server has not started to detach yet.
// The cap is short because the error can also mean that a server outside the configuration uses the storage.
func RetryOnStorageAttached[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return retryOnErrorCode(ctx, upcloud.ErrCodeStorageAttached, fn)
}

func retryOnErrorCode[T any](ctx context.Context, code string, fn func() (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute) // shortcut: fixed cap, add a timeouts block if users need longer
	defer cancel()
	for {
		res, err := fn()
		var p *upcloud.Problem
		if !errors.As(err, &p) || p.ErrorCode() != code {
			return res, err
		}
		select {
		case <-ctx.Done():
			return res, err
		case <-time.After(stateIllegalRetryInterval):
		}
	}
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

// From Kubernetes random suffixes.
const RandomSuffixChars = "bcdfghjklmnpqrstvwxz2456789"

func randomSuffix(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = RandomSuffixChars[rand.Intn(len(RandomSuffixChars))]
	}
	return string(b)
}

func WithRandomSuffix(input string) string {
	return input + "-" + randomSuffix(5)
}

func ValueInt32OrNil(v types.Int32) *int32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	s := v.ValueInt32()
	return &s
}

func ValueStringOrNil(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	s := v.ValueString()
	return &s
}

func ValueOrEmpty[T any](p *T) T {
	var v T
	if p != nil {
		v = *p
	}

	return v
}
