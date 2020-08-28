package upcloud

import (
	"regexp"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
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
