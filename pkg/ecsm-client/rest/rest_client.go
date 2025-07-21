package rest

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultAPIVersion = "v1"
	defaultAPIPath    = "api"
)

type Interface interface {
	Verb(verb string) *Request
	Get() *Request
	Put() *Request
	Post() *Request
	Delete() *Request
	APIVersion() string
}

// Client 是与 ECSM API Server 交互的客户端。
type RESTClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	apiVersion string
	apiPath    string
}

// NewClient 创建一个新的 ECSM 客户端实例。
func NewRESTClient(protocol, host, port string, httpClient *http.Client) (*RESTClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURLStr := fmt.Sprintf("%s://%s:%s", protocol, host, port)
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}

	return &RESTClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		apiVersion: defaultAPIVersion,
		apiPath:    defaultAPIPath,
	}, nil
}

func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

// Post begins a POST request. Short for c.Verb("POST").
func (c *RESTClient) Post() *Request {
	return c.Verb("POST")
}

// Put begins a PUT request. Short for c.Verb("PUT").
func (c *RESTClient) Put() *Request {
	return c.Verb("PUT")
}

// Get begins a GET request. Short for c.Verb("GET").
func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}

// Delete begins a DELETE request. Short for c.Verb("DELETE").
func (c *RESTClient) Delete() *Request {
	return c.Verb("DELETE")
}

// APIVersion returns the APIVersion this RESTClient is expected to use.
func (c *RESTClient) APIVersion() string {
	return fmt.Sprintf("%s/%s", c.apiPath, c.apiVersion)
}
