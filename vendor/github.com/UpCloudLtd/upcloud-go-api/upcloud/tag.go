package upcloud

// Tags represents a list of tags
type Tags struct {
	Tags []Tag `xml:"tag"`
}

// Tag represents a server tag
type Tag struct {
	Name        string   `xml:"name"`
	Description string   `xml:"description,omitempty"`
	Servers     []string `xml:"servers>server"`
}
