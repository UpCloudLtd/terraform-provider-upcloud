package utils

import (
	"os"
	"testing"
)

// ReadTestDataFile reads testdata from file to a string. Fails tests with Fatal, if reading the file fails.
func ReadTestDataFile(t *testing.T, name string) string {
	testdata, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(testdata)
}
