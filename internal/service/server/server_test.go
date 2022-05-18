package server

import (
	"fmt"
	"strings"
	"testing"
)

func TestServerDefaultTitle(t *testing.T) {
	longHostname := strings.Repeat("x", 255)
	suffixLength := 24
	want := fmt.Sprintf("%s… (managed by terraform)", longHostname[0:255-suffixLength])
	got := serverDefaultTitleFromHostname(longHostname)
	if want != got {
		t.Errorf("cloudServerDefaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}

	want = "terraform (managed by terraform)"
	got = serverDefaultTitleFromHostname("terraform")
	if want != got {
		t.Errorf("cloudServerDefaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}
}
