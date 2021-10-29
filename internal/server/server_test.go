package server

import "testing"

func TestBuildSimpleBackupOpts_basic(t *testing.T) {
	attrs := map[string]interface{}{
		"time": "2200",
		"plan": "weeklies",
	}

	sb := BuildSimpleBackupOpts(attrs)
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

	sb := BuildSimpleBackupOpts(attrs)
	expected := "no"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}
