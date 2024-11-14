package utils

import "fmt"

// DescriptionWithDeprecationWarning adds deprecation message to the description in a warning box before the actual description. Remember to also define DeprecationMessage for the resource/data-source.
func DescriptionWithDeprecationWarning(deprecationMessage, description string) string {
	return fmt.Sprintf(`~> %s

%s`, deprecationMessage, description)
}
