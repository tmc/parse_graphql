package parse

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Client is the primary struct that this package provides. It represents the
// connection to the Parse API
type Client struct {
	appID        string
	restApiKey   string
	masterKey    string
	sessionToken string

	logger *log.Logger
}

// NewClient creates a new Client to interact with the Parse API.
func NewClient(parseAppID string, RESTAPIKey string) (*Client, error) {
	return &Client{appID: parseAppID, restApiKey: RESTAPIKey}, nil
}

// WithMasterKey returns a Client with the master key set instead of the REST API key.
func (c *Client) WithMasterKey(masterKey string) *Client {
	newClient, _ := NewClient(c.appID, "")
	newClient.masterKey = masterKey
	newClient.logger = c.logger
	return newClient
}

// WithSessionToken returns a Client with the session token set, authenticating as the
// user associated with the token.
func (c *Client) WithSessionToken(sessionToken string) *Client {
	newClient, _ := NewClient(c.appID, c.restApiKey)
	newClient.sessionToken = sessionToken
	newClient.logger = c.logger
	return newClient
}

// TraceOn turns on API response tracing to the given logger.
func (c *Client) TraceOn(logger *log.Logger) {
	c.logger = logger
}

// TraceOff turns on API response tracing
func (c *Client) TraceOff() {
	c.logger = nil
}

func (c *Client) trace(args ...interface{}) {
	if c.logger != nil {
		c.logger.Println(args)
	}
}

func (c *Client) prepReq(method, url, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Parse-Application-ID", c.appID)
	if c.masterKey != "" {
		req.Header.Add("X-Parse-Master-Key", c.masterKey)
	} else {
		req.Header.Add("X-Parse-REST-API-Key", c.restApiKey)
	}
	if c.sessionToken != "" {
		req.Header.Add("X-Parse-Session-Token", c.sessionToken)
	}
	req.Header.Add("Content-Type", contentType)
	return req, err
}

func (c *Client) doSimple(method string, endpoint string) (*http.Response, error) {
	return c.do(method, endpoint, "application/json", nil)
}

func (c *Client) doWithBody(method string, endpoint string, body io.Reader) (*http.Response, error) {
	return c.do(method, endpoint, "application/json", body)
}

func (c *Client) do(method, endpoint, contentType string, body io.Reader) (*http.Response, error) {
	u, err := url.Parse(BaseURL + endpoint)
	if err != nil {
		return nil, err
	}
	req, err := c.prepReq(method, u.String(), contentType, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
	case 201:
	case 400:
		fallthrough
	case 404:
		defer resp.Body.Close()
		err, _ := unmarshalError(resp.Body)
		return nil, err
	case 401:
		return resp, ErrUnauthorized
	default:
		return resp, fmt.Errorf("got unexpected status code %d", resp.StatusCode)
	}
	return resp, err
}
