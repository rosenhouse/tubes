package mocks

import "encoding/json"

type JSONClient struct {
	GetCall struct {
		Receives struct {
			Path         string
			ResponseData interface{}
		}
		Returns struct {
			Error error
		}

		ResponseJSON string
	}
}

func (c *JSONClient) Get(path string, responseData interface{}) error {
	c.GetCall.Receives.Path = path
	c.GetCall.Receives.ResponseData = responseData

	json.Unmarshal([]byte(c.GetCall.ResponseJSON), responseData)

	return c.GetCall.Returns.Error
}
