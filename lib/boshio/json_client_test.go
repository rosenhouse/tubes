package boshio_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/boshio"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("JSON Client", func() {
	var (
		client         *boshio.JSONClient
		httpClient     *mocks.HTTPClient
		responseStruct struct {
			SomeField string `json:"SomeField"`
		}
	)

	BeforeEach(func() {
		httpClient = &mocks.HTTPClient{}
		client = &boshio.JSONClient{HTTPClient: httpClient}
		responseStruct.SomeField = ""
	})

	It("should make an HTTP Get call", func() {
		httpClient.GetCall.Returns.Body = []byte(`{ "SomeField": "some value" }`)
		Expect(client.Get("/some/path", &responseStruct)).To(Succeed())
		Expect(responseStruct.SomeField).To(Equal("some value"))
		Expect(httpClient.GetCall.Receives.Path).To(Equal("/some/path"))
	})

	Context("when the http client errors", func() {
		It("should return the error", func() {
			httpClient.GetCall.Returns.Error = errors.New("some error")
			Expect(client.Get("/some/path", &responseStruct)).To(MatchError("some error"))
		})
	})

	Context("when http client returns malformed JSON", func() {
		It("should return an error", func() {
			httpClient.GetCall.Returns.Body = []byte(`{`)
			err := client.Get("/some/path", &responseStruct)
			Expect(err).To(MatchError(HavePrefix("server returned malformed JSON")))
		})
	})

})
