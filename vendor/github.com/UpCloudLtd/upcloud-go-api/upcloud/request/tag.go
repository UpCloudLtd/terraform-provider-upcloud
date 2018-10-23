package request

import (
	"encoding/xml"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// CreateTagRequest represents a request to create a tag and assign it to zero or more servers
type CreateTagRequest struct {
	upcloud.Tag

	XMLName xml.Name `xml:"tag"`
}

// RequestURL implements the Request interface
func (r *CreateTagRequest) RequestURL() string {
	return "/tag"
}

// ModifyTagRequest represents a request to modify an existing tag. The Name is the name of the current tag, the Tag
// is the new values for the tag.
type ModifyTagRequest struct {
	upcloud.Tag

	XMLName xml.Name `xml:"tag"`
	Name    string   `xml:"-"`
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
