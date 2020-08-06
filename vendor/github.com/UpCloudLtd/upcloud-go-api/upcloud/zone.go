package upcloud

// Zones represents a /zone response
type Zones struct {
	Zones []Zone `xml:"zone"`
}

// Zone represents a zone
type Zone struct {
	Id          string `xml:"id"`
	Description string `xml:"description"`
}
