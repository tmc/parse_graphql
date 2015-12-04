package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Create creates a Parse object. On success the new object's ID is returned.
// The provided object is not modified.
func (c *Client) Create(object Object) (objectID string, err error) {
	className, err := objectTypeName(object)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(object)
	uri := "/1/classes/" + className
	resp, err := c.doWithBody("POST", uri, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	c.trace("Create", uri)
	_, id := objectURIToClassAndID(resp.Header.Get("Location"))
	return id, nil
}

// GetClass populates the passed object by looking up based on Class name and objectID.
func (c *Client) GetClass(className string, objectID string, object interface{}) error {
	uri := fmt.Sprintf("/1/classes/%s/%s", className, objectID)
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.trace("Get", uri, string(body))
	// TODO(tmc): warn if not == .Zero() before populating?
	return json.Unmarshal(body, object)
}

// Get populates the passed object by looking up based on objectID.
func (c *Client) Get(objectID string, object Object) error {
	className, err := objectTypeName(object)
	if err != nil {
		return err
	}
	return c.GetClass(className, objectID, object)
}

// Update submits the JSON serialization of object and on success returns the
// updated time. The provided object is not modified.
func (c *Client) Update(object Object) (updateTime time.Time, err error) {
	className, err := objectTypeName(object)
	if err != nil {
		return updateTime, err
	}
	payload, err := json.Marshal(object)
	uri := fmt.Sprintf("/1/classes/%s/%s", className, object.ObjectID())
	resp, err := c.doWithBody("PUT", uri, bytes.NewReader(payload))
	if err != nil {
		return updateTime, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return updateTime, err
	}
	c.trace("Update", uri, string(body))
	updatedAt := &struct {
		Time time.Time `json:"updatedAt"`
	}{}
	err = json.Unmarshal(body, updatedAt)
	return updatedAt.Time, err
}

// Delete removes the provided object from the Parse data store.
func (c *Client) Delete(object Object) error {
	className, err := objectTypeName(object)
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("/1/classes/%s/%s", className, object.ObjectID())
	resp, err := c.doSimple("DELETE", uri)
	defer resp.Body.Close()
	c.trace("Delete", uri)
	return err
}
