package request

// Request is the interface for request objects
type Request interface {
	// RequestURL returns the relative API URL for the request, excluding the API version.
	RequestURL() string
}
