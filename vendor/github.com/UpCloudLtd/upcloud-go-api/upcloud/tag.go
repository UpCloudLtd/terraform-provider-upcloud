package upcloud

import (
	"encoding/json"
)

// Tags represents a list of tags
type Tags struct {
	Tags []Tag
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Tags) UnmarshalJSON(b []byte) error {
	type localTag Tag
	type tagWrapper struct {
		Tags []localTag `json:"tag"`
	}

	v := struct {
		Tags tagWrapper `json:"tags"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	for _, t := range v.Tags.Tags {
		s.Tags = append(s.Tags, Tag(t))
	}

	return nil
}

// TagServerSlice is a slice of string.
// It exists to allow for a custom JSON unmarshaller.
type TagServerSlice []string

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (t *TagServerSlice) UnmarshalJSON(b []byte) error {
	v := struct {
		Servers []string `json:"server"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*t) = v.Servers

	return nil
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (t TagServerSlice) MarshalJSON() ([]byte, error) {
	v := struct {
		Servers []string `json:"server"`
	}{}
	if t == nil {
		t = make(TagServerSlice, 0)
	}
	v.Servers = t

	return json.Marshal(v)
}

// Tag represents a server tag
type Tag struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Servers     TagServerSlice `json:"servers"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (s *Tag) UnmarshalJSON(b []byte) error {
	type localTag Tag

	v := struct {
		Tag localTag `json:"tag"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	(*s) = Tag(v.Tag)

	return nil
}
