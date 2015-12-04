package parse

import (
	"encoding/json"
	"io"
)

// ParseFile is a Parse File that has been uploaded.
type ParseFile struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// UploadFile uploads a Parse File from the provided filename, contents and
// content type.
func (c *Client) UploadFile(name string, contents io.Reader, contentType string) (*ParseFile, error) {
	uri := "/1/files/"+name
	resp, err := c.do("POST", uri, contentType, contents)
	c.trace("UploadFile", uri, contentType)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result *ParseFile
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// DeleteFile removes a Parse File.
func (c *Client) DeleteFile(fileName string) error {
	if c.masterKey == "" {
		return ErrRequiresMasterKey
	}
	resp, err := c.doSimple("DELETE", "/1/files/"+fileName)
	if err != nil {
		return err
	}
	c.trace("DeleteFile", fileName)
	defer resp.Body.Close()
	return nil
}
