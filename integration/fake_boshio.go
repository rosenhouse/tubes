package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type FakeBoshIO struct{}

func (f *FakeBoshIO) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent":
		f.handleStemcell(w, r)
	case "/stemcell-download":
		w.Write([]byte("some stemcell bytes"))
	case "/api/v1/releases/github.com/cloudfoundry-incubator/bosh-aws-cpi-release":
		f.serveReleaseInfo("aws-cpi-release", "1234", "some-aws-cpi-release-download-url", w)
	case "/releases/github.com/cloudfoundry-incubator/bosh-aws-cpi-release":
		w.Write([]byte(`sha1: 22596363b3de40b06f981fb85d82312e8c0ed511`))
	case "/api/v1/releases/github.com/cloudfoundry/bosh":
		f.serveReleaseInfo("bosh-release", "5678", "some-bosh-release-download-url", w)
	case "/releases/github.com/cloudfoundry/bosh":
		w.Write([]byte(`sha1: 52e98718f012ca15f876ae405b57848b5c7128dd`))
	default:
		fmt.Printf("\n\t bosh.io server got request: %+v\n", r)
		w.WriteHeader(http.StatusTeapot)
	}
}

func (f *FakeBoshIO) handleStemcell(w http.ResponseWriter, r *http.Request) {
	responseData := make([]struct {
		Light struct {
			URL string
		}
	}, 1)
	responseData[0].Light.URL = "/stemcell-download"
	respBytes, _ := json.Marshal(responseData)
	w.Write(respBytes)
}

func (f *FakeBoshIO) serveReleaseInfo(name, version, downloadURL string, w http.ResponseWriter) {
	responseData := make([]struct {
		Name    string
		Version string
		URL     string
	}, 1)
	responseData[0].Name = name
	responseData[0].Version = version
	responseData[0].URL = downloadURL

	respBytes, _ := json.Marshal(responseData)
	w.Write(respBytes)
}
