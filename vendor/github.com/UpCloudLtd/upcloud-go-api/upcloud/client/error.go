package client

import "fmt"

// Error represents an error returned from the client. Errors are thrown when requests don't have a successful status code
type Error struct {
	ErrorCode    int
	ErrorMessage string
	ResponseBody []byte
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.ErrorCode, e.ErrorMessage)
}
