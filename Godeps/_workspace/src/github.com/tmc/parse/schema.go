package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type SchemaField struct {
	Type        string `json:"type,omitempty"`
	TargetClass string `json:"targetClass,omitempty"`
}

type Schema struct {
	ClassName string                 `json:"className,omitempty"`
	Fields    map[string]SchemaField `json:"fields,omitempty"`
}

//
func (c *Client) GetClassSchema(className string) (*Schema, error) {
	uri := fmt.Sprintf("/1/schemas/%s", className)
	resp, err := c.doSimple("GET", uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var schema *Schema
	c.trace("GetClassSchema", uri, string(body))
	return schema, json.Unmarshal(body, &schema)
}

func (c *Client) GetFullSchema() (map[string]*Schema, error) {
	uri := fmt.Sprintf("/1/schemas/")
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
		Schemas []*Schema `json:"results"`
	}
	c.trace("GetFullSchema", uri, string(body))
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	mapResult := make(map[string]*Schema)
	for _, s := range result.Schemas {
		mapResult[s.ClassName] = s
	}
	return mapResult, nil
}
