package client

import (
	"bytes"
	"fmt"
	"github.com/blang/semver"
	"github.com/hashicorp/go-cleanhttp"
	"io"
	"io/ioutil"
	"net/http"
	"time"
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
	apiBaseUrl string
}

// New creates ands returns a new client configured with the specified user and password
func New(userName, password string) *Client {
	client := Client{}

	client.userName = userName
	client.password = password
	client.httpClient = cleanhttp.DefaultClient()
	client.SetTimeout(time.Second * DEFAULT_TIMEOUT)

	client.apiVersion = DEFAULT_API_VERSION
	client.apiBaseUrl = DEFAULT_API_BASEURL

	return &client
}

// SetTimeout sets the client timeout to the specified amount of seconds
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// CreateRequestUrl creates and returns a complete request URL for the specified API location
func (c *Client) CreateRequestUrl(location string) string {
	return fmt.Sprintf("%s%s", c.getBaseUrl(), location)
}

// PerformGetRequest performs a GET request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformGetRequest(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	return c.performRequest(request)
}

// PerformPostRequest performs a POST request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformPostRequest(url string, requestBody []byte) ([]byte, error) {
	var bodyReader io.Reader

	if requestBody != nil {
		bodyReader = bytes.NewBuffer(requestBody)
	}

	request, err := http.NewRequest("POST", url, bodyReader)

	if err != nil {
		return nil, err
	}

	return c.performRequest(request)
}

// PerformPutRequest performs a PUT request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformPutRequest(url string, requestBody []byte) ([]byte, error) {
	var bodyReader io.Reader

	if requestBody != nil {
		bodyReader = bytes.NewBuffer(requestBody)
	}

	request, err := http.NewRequest("PUT", url, bodyReader)

	if err != nil {
		return nil, err
	}

	return c.performRequest(request)
}

// PerformDeleteRequest performs a DELETE request to the specified URL and returns the response body and eventual errors
func (c *Client) PerformDeleteRequest(url string) error {
	request, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		return err
	}

	_, err = c.performRequest(request)
	return err
}

// Adds common headers to the specified request
func (c *Client) addRequestHeaders(request *http.Request) *http.Request {
	request.SetBasicAuth(c.userName, c.password)
	request.Header.Add("Accept", "application/xml")
	request.Header.Add("Content-Type", "application/xml")

	return request
}

// Performs the specified HTTP request and returns the response through handleResponse()
func (c *Client) performRequest(request *http.Request) ([]byte, error) {
	c.addRequestHeaders(request)
	response, err := c.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	return handleResponse(response)
}

// Returns the base URL to use for API requests
func (c *Client) getBaseUrl() string {
	urlVersion, _ := semver.Make(c.apiVersion)

	return fmt.Sprintf("%s/%d.%d", c.apiBaseUrl, urlVersion.Major, urlVersion.Minor)
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
