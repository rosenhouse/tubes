package boshio

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/rosenhouse/tubes/lib/director"
)

type jsonClient interface {
	Get(route string, responseData interface{}) error
}

type Client struct {
	JSONClient jsonClient
	HTTPClient httpClient
}

func (c *Client) LatestRelease(releasePath string) (director.Artifact, error) {
	var artifact director.Artifact
	var responseData []struct {
		Name    string
		Version string
		URL     string
	}
	url := "/api/v1/releases/" + releasePath
	err := c.JSONClient.Get(url, &responseData)
	if err != nil {
		return artifact, err
	}
	if len(responseData) == 0 {
		return artifact, fmt.Errorf("empty result for %s", url)
	}
	artifact.URL = responseData[0].URL

	htmlURL := fmt.Sprintf("/releases/%s?version=%s", releasePath, responseData[0].Version)

	respBytes, err := c.HTTPClient.Get(htmlURL)
	if err != nil {
		return artifact, err
	}

	const sha1Pattern = `\w*sha1\: ([a-f0-9]{40})`
	regex := regexp.MustCompile(sha1Pattern)
	matches := regex.FindSubmatch(respBytes)
	if len(matches) != 2 {
		return artifact, fmt.Errorf(
			"failed while scraping %s: unable to find sha1", htmlURL)
	}
	artifact.SHA = string(matches[1])
	return artifact, nil
}

func (c *Client) LatestStemcell(stemcellName string) (director.Artifact, error) {
	var artifact director.Artifact
	var responseData []struct {
		Light struct {
			Size int
			URL  string
		}
	}
	url := "/api/v1/stemcells/" + stemcellName
	err := c.JSONClient.Get(url, &responseData)
	if err != nil {
		return artifact, err
	}
	if len(responseData) == 0 {
		return artifact, fmt.Errorf("empty result for %s", url)
	}
	artifact.URL = responseData[0].Light.URL
	stemcellBytes, err := c.HTTPClient.Get(artifact.URL)
	if err != nil {
		return artifact, fmt.Errorf("error while downloading stemcell: %s", err)
	}
	sha1Bytes := sha1.Sum(stemcellBytes)
	artifact.SHA = hex.EncodeToString(sha1Bytes[:])
	return artifact, nil
}
