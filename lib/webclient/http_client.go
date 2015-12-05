package webclient

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HTTPClient struct {
	BaseURL       string
	SkipTLSVerify bool
}

func (c *HTTPClient) resolvePath(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}

func (c *HTTPClient) Get(path string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.SkipTLSVerify},
	}
	client := &http.Client{Transport: tr}

	fullURL, err := c.resolvePath(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err // not tested
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned status code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
