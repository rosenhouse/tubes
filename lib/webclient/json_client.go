package webclient

import (
	"encoding/json"
	"fmt"
)

type httpClient interface {
	Get(path string) ([]byte, error)
}

type JSONClient struct {
	HTTPClient httpClient
}

func (c *JSONClient) Get(route string, responseData interface{}) error {
	responseBody, err := c.HTTPClient.Get(route)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBody, &responseData)
	if err != nil {
		return fmt.Errorf("server returned malformed JSON: %s", err)
	}
	return nil
}
