package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// CallCloudFunction invokes the given cloud code function.
// The arguments parameter is serialized as JSON and provided as parameters.
func (c *Client) CallCloudFunction(functionName string, arguments map[string]interface{}) ([]byte, error) {
	if arguments == nil {
		arguments = map[string]interface{}{}
	}
	payload, err := json.Marshal(arguments)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("/1/functions/%s", functionName)
	resp, err := c.doWithBody("POST", uri, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.trace("CallCloudFunction err", uri, string(body))
		return nil, err
	}
	c.trace("CallCloudFunction", uri, string(body))
	return body, err
}

// CallCloudJob schedules the given cloud code job.
// The arguments parameter is serialized as JSON and provided as parameters.
func (c *Client) CallCloudJob(jobName string, arguments interface{}) ([]byte, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("/1/jobs/%s", jobName)
	resp, err := c.doWithBody("POST", uri, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.trace("CallCloudJob err", uri, string(body))
		return nil, err
	}
	c.trace("CallCloudJob", uri, string(body))
	return body, err
}
