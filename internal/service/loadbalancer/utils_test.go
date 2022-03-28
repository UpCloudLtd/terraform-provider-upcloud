package loadbalancer

import "testing"

func TestMarshalID(t *testing.T) {
	want := "load/balancer/frontend"
	got := marshalID("load", "balancer", "frontend")
	if want != got {
		t.Errorf("marshalID failed want %s got %s", want, got)
	}
}

func TestUnmarshalID(t *testing.T) {
	var load, balancer, frontend string
	id := "load/balancer/frontend"
	err := unmarshalID(id, &load, &balancer, &frontend)
	if err != nil {
		t.Fatal(err)
	}
	if load != "load" {
		t.Errorf("unmarshalID failed want load got %s", load)
	}
	if balancer != "balancer" {
		t.Errorf("unmarshalID failed want balancer got %s", balancer)
	}
	if frontend != "frontend" {
		t.Errorf("unmarshalID failed want frontend got %s", frontend)
	}
	err = unmarshalID(id, &load, &balancer)
	if err == nil {
		t.Fatal("unmarshalID failed expected 'not enough components' error got nil")
	}
}
