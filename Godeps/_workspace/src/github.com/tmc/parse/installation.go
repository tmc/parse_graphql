package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"
)

// Installation is the minimal interface a type reprenting an installation must satisfy.
type Installation interface {
	Object
}

// ParseInstallation is a representation of an application installed on a device.
type ParseInstallation struct {
	ParseObject
	Channels    []string `json:"channels,omitempty"`
	DeviceToken string   `json:"deviceToken,omitempty"`
	DeviceType  string   `json:"deviceType,omitempty"`
}

// CreateInstallation creates an installation from an Installation object.
func (c *Client) CreateInstallation(installation Installation) (installationID string, err error) {
	payload, err := json.Marshal(installation)
	resp, err := c.doWithBody("POST", "/1/installations", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var i ParseInstallation
	err = json.Unmarshal(body, &i)
	c.trace("CreateInstallation", "/1/installations", string(body))
	return i.ID, err
}

// GetInstallation populates the provided installation based on the installationID.
func (c *Client) GetInstallation(installationID string, installation Installation) error {
	uri := fmt.Sprintf("/1/installations/%s", installationID)
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.trace("GetInstallation", uri, string(body))
	// TODO(tmc): warn if not == .Zero() before populating?
	return json.Unmarshal(body, installation)
}

// UpdateInstallation updates an installation. The provided installation is not modified.
// On success the updated time is returned.
func (c *Client) UpdateInstallation(installation Installation) (updateTime time.Time, err error) {
	payload, err := json.Marshal(installation)
	uri := fmt.Sprintf("/1/installations/%s", installation.ObjectID())
	resp, err := c.doWithBody("PUT", uri, bytes.NewReader(payload))
	if err != nil {
		return updateTime, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return updateTime, err
	}
	c.trace("UpdateInstallation", uri, string(body))
	updatedAt := &struct {
		Time time.Time `json:"updatedAt"`
	}{}
	err = json.Unmarshal(body, updatedAt)
	return updatedAt.Time, err
}

// DeleteInstallation removes an installation by ID.
func (c *Client) DeleteInstallation(installation Installation) error {
	uri := fmt.Sprintf("/1/installations/%s", installation.ObjectID())
	resp, err := c.doSimple("DELETE", uri)
	defer resp.Body.Close()
	c.trace("DeleteInstallation", uri)
	return err
}

// QueryInstallations queries Installation objects based on the provided options.
func (c *Client) QueryInstallations(options *QueryOptions, destination []Installation) error {
	uri, err := url.Parse("/1/installations")

	if options != nil {
		params := uri.Query()
		if options.Where != "" {
			params.Set("where", options.Where)
		}
		uri.RawQuery = params.Encode()
	}

	resp, err := c.doSimple("GET", uri.String())
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// delay parsing of results
	c.trace("Query", uri, string(body))
	results := struct {
		Results json.RawMessage `json:"results"`
	}{}
	// first pass
	err = json.Unmarshal(body, &results)
	if err != nil {
		return err
	}
	return json.Unmarshal(results.Results, destination)
}
