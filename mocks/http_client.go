package mocks

type HTTPClient struct {
	GetCall struct {
		Receives struct {
			Path string
		}
		Returns struct {
			Body  []byte
			Error error
		}
	}
}

func (c *HTTPClient) Get(path string) ([]byte, error) {
	c.GetCall.Receives.Path = path
	return c.GetCall.Returns.Body, c.GetCall.Returns.Error
}
