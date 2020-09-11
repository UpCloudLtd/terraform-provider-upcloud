package upcloud

import (
	"encoding/json"
	"fmt"
)

// Error represents an error
type Error struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (e *Error) UnmarshalJSON(b []byte) error {
	type localError Error
	v := struct {
		Error localError `json:"error"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*e) = Error(v.Error)

	return nil
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.ErrorMessage, e.ErrorCode)
}
