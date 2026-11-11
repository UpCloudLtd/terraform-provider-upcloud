package utils

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"

	"github.com/stretchr/testify/assert"
)

func TestFilterNetworks(t *testing.T) {
	toFilter := []upcloud.Network{
		{
			Name: "aa",
		},
		{
			Name: "ba",
		},
		{
			Name: "ca",
		},
		{
			Name: "ab",
		},
		{
			Name: "ac",
		},
		{
			Name: "bda",
		},
		{
			Name: "bdb",
		},
		{
			Name: "bdc",
		},
	}

	filtered, err := FilterNetworks(toFilter, func(n upcloud.Network) (bool, error) {
		return regexp.MatchString("^a.*", n.Name)
	})
	if err != nil {
		t.Log("filter returned error")
		t.Fail()
	}
	if len(filtered) != 3 {
		t.Logf("filter returned wrong number of items: %d", len(filtered))
		t.Fail()
	}

	filtered2, err := FilterNetworks(toFilter, func(n upcloud.Network) (bool, error) {
		return regexp.MatchString("^.*d.*$", n.Name)
	})
	if err != nil {
		t.Log("filter2 returned error")
		t.Fail()
	}
	if len(filtered2) != 3 {
		t.Logf("filter2 returned wrong number of items: %d", len(filtered2))
		t.Fail()
	}
}

func TestWithRetry(t *testing.T) {
	fail := func() (interface{}, error) {
		return nil, fmt.Errorf("")
	}

	count := 0
	successAftertree := func() (interface{}, error) {
		if count < 3 {
			count++
			return nil, fmt.Errorf("")
		}

		return nil, nil
	}

	if _, err := WithRetry(fail, 3, 0); err == nil {
		t.Log("should fail")
		t.Fail()
	}
	if _, err := WithRetry(successAftertree, 4, 0); err != nil {
		t.Log("should not fail")
		t.Fail()
	}
	count = 0
	if _, err := WithRetry(successAftertree, 3, 0); err == nil {
		t.Log("should fail")
		t.Fail()
	}
}

func TestRetryOnServerStateIllegal(t *testing.T) {
	defer func(d time.Duration) { stateIllegalRetryInterval = d }(stateIllegalRetryInterval)
	stateIllegalRetryInterval = 0
	problem := func(code string) error {
		return &upcloud.Problem{Type: code, Status: http.StatusConflict}
	}

	calls := 0
	res, err := RetryOnServerStateIllegal(context.Background(), func() (int, error) {
		calls++
		if calls < 3 {
			return 0, problem(upcloud.ErrCodeServerStateIllegal)
		}
		return 42, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 42, res)
	assert.Equal(t, 3, calls)

	// Other errors are returned without retrying.
	calls = 0
	_, err = RetryOnServerStateIllegal(context.Background(), func() (int, error) {
		calls++
		return 0, problem(upcloud.ErrCodeResourceAlreadyExists)
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)

	// Stops retrying when context is done and returns the last error.
	stateIllegalRetryInterval = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls = 0
	_, err = RetryOnServerStateIllegal(ctx, func() (int, error) {
		calls++
		return 0, problem(upcloud.ErrCodeServerStateIllegal)
	})
	var p *upcloud.Problem
	assert.ErrorAs(t, err, &p)
	assert.Equal(t, upcloud.ErrCodeServerStateIllegal, p.ErrorCode())
	assert.Equal(t, 1, calls)
}

func TestRetryOnStorageStateIllegal(t *testing.T) {
	defer func(d time.Duration) { stateIllegalRetryInterval = d }(stateIllegalRetryInterval)
	stateIllegalRetryInterval = 0

	calls := 0
	_, err := RetryOnStorageStateIllegal(context.Background(), func() (int, error) {
		calls++
		if calls < 3 {
			return 0, &upcloud.Problem{Type: upcloud.ErrCodeStorageStateIllegal, Status: http.StatusConflict}
		}
		return 0, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, calls)

	// SERVER_STATE_ILLEGAL is not retried here.
	calls = 0
	_, err = RetryOnStorageStateIllegal(context.Background(), func() (int, error) {
		calls++
		return 0, &upcloud.Problem{Type: upcloud.ErrCodeServerStateIllegal, Status: http.StatusConflict}
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestStorageAddressFormat(t *testing.T) {
	storageAddressWithAddress := "virtio:1"
	storageAddressWithoutAddress := "scsi"
	storageAddressEmpty := ""

	ret := StorageAddressFormat(storageAddressWithAddress)
	assert.Equal(t, ret, "virtio")

	ret = StorageAddressFormat(storageAddressWithoutAddress)
	assert.Equal(t, ret, "scsi")

	ret = StorageAddressFormat(storageAddressEmpty)
	assert.Equal(t, ret, "")
}

func TestMarshalID(t *testing.T) {
	want := "load/balancer/frontend"
	got := MarshalID("load", "balancer", "frontend")
	if want != got {
		t.Errorf("marshalID failed want %s got %s", want, got)
	}
}

func TestUnmarshalID(t *testing.T) {
	var load, balancer, frontend string
	id := "load/balancer/frontend"
	err := UnmarshalID(id, &load, &balancer, &frontend)
	if err != nil {
		t.Fatal(err)
	}
	if load != "load" {
		t.Errorf("utils.UnmarshalID failed want load got %s", load)
	}
	if balancer != "balancer" {
		t.Errorf("utils.UnmarshalID failed want balancer got %s", balancer)
	}
	if frontend != "frontend" {
		t.Errorf("utils.UnmarshalID failed want frontend got %s", frontend)
	}
	err = UnmarshalID(id, &load, &balancer)
	if err == nil {
		t.Fatal("utils.UnmarshalID failed expected 'not enough components' error got nil")
	}
}

func TestValueOrEmpty(t *testing.T) {
	var stringPtr *string
	assert.Equal(t, "", ValueOrEmpty(stringPtr))

	stringVal := "test"
	stringPtr = &stringVal
	assert.Equal(t, "test", ValueOrEmpty(stringPtr))

	var intPtr *int
	assert.Equal(t, 0, ValueOrEmpty(intPtr))

	intVal := 42
	intPtr = &intVal
	assert.Equal(t, 42, ValueOrEmpty(intPtr))
}
