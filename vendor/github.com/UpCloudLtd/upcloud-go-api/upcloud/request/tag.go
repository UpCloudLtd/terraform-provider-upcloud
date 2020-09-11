package request

import (
	"encoding/json"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// CreateTagRequest represents a request to create a tag and assign it to zero or more servers
type CreateTagRequest struct {
	upcloud.Tag
}

// RequestURL implements the Request interface
func (r *CreateTagRequest) RequestURL() string {
	return "/tag"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateTagRequest) MarshalJSON() ([]byte, error) {
	type localCreateTagRequest CreateTagRequest
	v := struct {
		CreateTagRequest localCreateTagRequest `json:"tag"`
	}{}
	v.CreateTagRequest = localCreateTagRequest(r)

	return json.Marshal(&v)
}

// ModifyTagRequest represents a request to modify an existing tag. The Name is the name of the current tag, the Tag
// is the new values for the tag.
type ModifyTagRequest struct {
	upcloud.Tag

	Name string `json:"-"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyTagRequest) MarshalJSON() ([]byte, error) {
	type localModifyTagRequest ModifyTagRequest
	v := struct {
		ModifyTagRequest localModifyTagRequest `json:"tag"`
	}{}
	v.ModifyTagRequest = localModifyTagRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *ModifyTagRequest) RequestURL() string {
	return fmt.Sprintf("/tag/%s", r.Name)
}

// DeleteTagRequest represents a request to delete a tag
type DeleteTagRequest struct {
	Name string
}

// RequestURL implements the Request interface
func (r *DeleteTagRequest) RequestURL() string {
	return fmt.Sprintf("/tag/%s", r.Name)
}
