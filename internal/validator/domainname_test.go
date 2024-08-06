package validator

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateDomainName(t *testing.T) {
	labelMaxLen := 63
	validNames := []string{
		"example",
		"Example",
		"0example",
		"example.com",
		"example.co.uk",
		"0.0.0.0.com",
		"a",
		strings.Repeat("a", labelMaxLen),
		fmt.Sprintf("%s.%s.%s.%s", strings.Repeat("a", labelMaxLen), strings.Repeat("b", labelMaxLen), strings.Repeat("c", labelMaxLen), strings.Repeat("d", labelMaxLen-2)),
		fmt.Sprintf("example.%s.%s.com", strings.Repeat("e", labelMaxLen), strings.Repeat("e", labelMaxLen)),
	}

	invalidNames := []string{
		".example",
		".-example",
		"0",
		"0.0.0.0",
		"example..com",
		"example.",
		"example.com.",
		"bücher.tld",
		".",
		strings.Repeat("a", labelMaxLen+1),
		fmt.Sprintf("example.%s.com", strings.Repeat("a", labelMaxLen+1)),
		fmt.Sprintf("%s.%s.%s.%s", strings.Repeat("a", labelMaxLen), strings.Repeat("b", labelMaxLen), strings.Repeat("c", labelMaxLen), strings.Repeat("d", labelMaxLen)),
	}

	for _, name := range validNames {
		if err := ValidateDomainName(name); err != nil {
			t.Errorf("serverValidateHostname failed '%s' is valid name: %s", name, err)
		}
	}

	for _, name := range invalidNames {
		err := ValidateDomainName(name)
		if err == nil {
			t.Errorf("serverValidateHostname failed '%s' is not a valid name", name)
		}
	}
}
