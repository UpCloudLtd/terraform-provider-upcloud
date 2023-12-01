package utils

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"

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
