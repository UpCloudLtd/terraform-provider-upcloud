package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerDefaultTitle(t *testing.T) {
	longHostname := strings.Repeat("x", 255)
	suffixLength := 24
	want := fmt.Sprintf("%s… (managed by terraform)", longHostname[0:255-suffixLength])
	got := defaultTitleFromHostname(longHostname)
	if want != got {
		t.Errorf("defaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}

	want = "terraform (managed by terraform)"
	got = defaultTitleFromHostname("terraform")
	if want != got {
		t.Errorf("defaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}
}

func TestBuildSimpleBackupOpts_basic(t *testing.T) {
	value := simpleBackupModel{
		Time: types.StringValue("2200"),
		Plan: types.StringValue("weeklies"),
	}

	sb := buildSimpleBackupOpts(&value)
	expected := "2200,weeklies"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}

func TestBuildSimpleBackupOpts_withInvalidInput(t *testing.T) {
	value := simpleBackupModel{
		Time: types.StringValue("2200"),
	}

	sb := buildSimpleBackupOpts(&value)
	expected := "no"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}
