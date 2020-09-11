package request

import (
	"encoding/json"
	"fmt"
)

// GetHostDetailsRequest represents the request for the details of a
// single private host
type GetHostDetailsRequest struct {
	ID int
}

// RequestURL implements the Request interface
func (r *GetHostDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/host/%d", r.ID)
}

// ModifyHostRequest represents the request to modify a private host
type ModifyHostRequest struct {
	ID          int    `json:"-"`
	Description string `json:"description"`
}

// RequestURL implements the Request interface
func (r *ModifyHostRequest) RequestURL() string {
	return fmt.Sprintf("/host/%d", r.ID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyHostRequest) MarshalJSON() ([]byte, error) {
	type localModifyHostRequest ModifyHostRequest
	v := struct {
		ModifyHostRequest localModifyHostRequest `json:"host"`
	}{}
	v.ModifyHostRequest = localModifyHostRequest(r)

	return json.Marshal(&v)
}
