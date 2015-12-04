package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type HookFunction struct {
	FunctionName string `json:"functionName,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (c *Client) GetHookFunctions() ([]*HookFunction, error) {
	uri := fmt.Sprintf("/1/hooks/functions")
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result struct {
		HookFunctions []*HookFunction `json:"results"`
	}
	c.trace("GetHookFunctions", uri, string(body))
	return result.HookFunctions, json.Unmarshal(body, &result)
}

func (c *Client) CreateHookFunction(fn *HookFunction) error {
	payload, err := json.Marshal(fn)
	c.trace("CreateHookFunction >", "/1/hooks/functions", string(payload))
	resp, err := c.doWithBody("POST", "/1/hooks/functions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response interface{}
	err = json.Unmarshal(body, &response)
	c.trace("CreateUser <", "/1/hooks/functions", string(body))
	return err
}
