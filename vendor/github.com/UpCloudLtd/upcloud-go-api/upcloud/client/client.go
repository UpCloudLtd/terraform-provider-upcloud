package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/hashicorp/go-cleanhttp"
)

// Constants
const (
	DEFAULT_API_VERSION = "1.2.3"
	DEFAULT_API_BASEURL = "https://api.upcloud.com"

	// The default timeout (in seconds)
	DEFAULT_TIMEOUT = 10
)

// Client represents an API client
type Client struct {
	userName   string
	password   string
	httpClient *http.Client

	apiVersion string
	apiBaseURL string
}

// New creates ands returns a new client configured with the specified user and password
func New(userName, password string) *Client {

	return NewWithHTTPClient(userName, password, cleanhttp.DefaultClient())
}

// NewWithHTTPClient creates ands returns a new client configured with the specified user and password and
// using a supplied `http.Client`.
func NewWithHTTPClient(userName string, password string, httpClient *http.Client) *Client {
	client := Client{}

	client.userName = userName
	client.password = password
	client.httpClient = httpClient
	client.SetTimeout(time.Second * DEFAULT_TIMEOUT)

	client.apiVersion = DEFAULT_API_VERSION
	client.apiBaseURL = DEFAULT_API_BASEURL

	return &client
}

// SetTimeout sets the client timeout to the specified amount of seconds
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// GetTimeout returns current timeout
func (c *Client) GetTimeout() time.Duration {
	return c.httpClient.Timeout
}

// CreateRequestURL creates and returns a complete request URL for the specified API location
func (c *Client) CreateRequestURL(location string) string {
	return fmt.Sprintf("%s%s", c.getBaseURL(), location)
}

// PerformJSONGetRequest performs a GET request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformJSONGetRequest(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	return c.performJSONRequest(request)
}

// PerformJSONPostRequest performs a POST request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformJSONPostRequest(url string, requestBody []byte) ([]byte, error) {
	var bodyReader io.Reader

	if requestBody != nil {
		bodyReader = bytes.NewBuffer(requestBody)
	}

	request, err := http.NewRequest(http.MethodPost, url, bodyReader)

	if err != nil {
		return nil, err
	}

	return c.performJSONRequest(request)
}

// PerformJSONPutRequest performs a PUT request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformJSONPutRequest(url string, requestBody []byte) ([]byte, error) {
	var bodyReader io.Reader

	if requestBody != nil {
		bodyReader = bytes.NewBuffer(requestBody)
	}

	request, err := http.NewRequest(http.MethodPut, url, bodyReader)

	if err != nil {
		return nil, err
	}

	return c.performJSONRequest(request)
}

// PerformJSONDeleteRequest performs a DELETE request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformJSONDeleteRequest(url string) error {
	request, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		return err
	}

	_, err = c.performJSONRequest(request)
	return err
}

// Adds common headers to the specified request
func (c *Client) addJSONRequestHeaders(request *http.Request) *http.Request {
	request.SetBasicAuth(c.userName, c.password)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	return request
}

// Performs the specified HTTP request and returns the response through handleResponse()
func (c *Client) performJSONRequest(request *http.Request) ([]byte, error) {
	c.addJSONRequestHeaders(request)
	response, err := c.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	return handleResponse(response)
}

// Returns the base URL to use for API requests
func (c *Client) getBaseURL() string {
	urlVersion, _ := semver.Make(c.apiVersion)

	return fmt.Sprintf("%s/%d.%d", c.apiBaseURL, urlVersion.Major, urlVersion.Minor)
}

// Parses the response and returns either the response body or an error
func handleResponse(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	// Return an error on unsuccessful requests
	if response.StatusCode < 200 || response.StatusCode > 299 {
		errorBody, _ := ioutil.ReadAll(response.Body)

		return nil, &Error{response.StatusCode, response.Status, errorBody}
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	return responseBody, err
}
