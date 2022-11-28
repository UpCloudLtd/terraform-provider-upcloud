package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/stretchr/testify/assert"
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
	attrs := map[string]interface{}{
		"time": "2200",
		"plan": "weeklies",
	}

	sb := buildSimpleBackupOpts(attrs)
	expected := "2200,weeklies"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}

func TestBuildSimpleBackupOpts_withInvalidInput(t *testing.T) {
	attrs := map[string]interface{}{
		"time":     "2200",
		"interval": "daily",
		"retetion": 7,
	}

	sb := buildSimpleBackupOpts(attrs)
	expected := "no"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}

func TestBuildLabels(t *testing.T) {
	attr := map[string]interface{}{
		"origin": "unit-test",
	}

	l := buildLabels(attr)
	assert.Equal(t, &upcloud.LabelSlice{upcloud.Label{Key: "origin", Value: "unit-test"}}, l)
}
