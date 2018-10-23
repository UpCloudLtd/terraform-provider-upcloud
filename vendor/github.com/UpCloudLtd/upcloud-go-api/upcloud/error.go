package upcloud

import "fmt"

// Error represents an error
type Error struct {
	ErrorCode    string `xml:"error_code"`
	ErrorMessage string `xml:"error_message"`
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.ErrorMessage, e.ErrorCode)
}
