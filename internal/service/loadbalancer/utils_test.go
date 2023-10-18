package loadbalancer

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
)

func TestMarshalID(t *testing.T) {
	want := "load/balancer/frontend"
	got := utils.MarshalID("load", "balancer", "frontend")
	if want != got {
		t.Errorf("marshalID failed want %s got %s", want, got)
	}
}

func TestUnmarshalID(t *testing.T) {
	var load, balancer, frontend string
	id := "load/balancer/frontend"
	err := utils.UnmarshalID(id, &load, &balancer, &frontend)
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
	err = utils.UnmarshalID(id, &load, &balancer)
	if err == nil {
		t.Fatal("utils.UnmarshalID failed expected 'not enough components' error got nil")
	}
}
