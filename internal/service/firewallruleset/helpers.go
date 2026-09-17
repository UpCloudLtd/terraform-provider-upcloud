package firewallruleset

import (
	"fmt"
)

func interfaceString(v interface{}) string {
	if v == nil {
		return ""
	}

	s, ok := v.(string)
	if ok {
		return s
	}

	return fmt.Sprint(v)
}
