package boshio_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/boshio"
)

var _ = Describe("HTTP Client", func() {
	var (
		server     *httptest.Server
		c          *boshio.HTTPClient
		requestURL *url.URL
	)

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURL = r.URL
			w.Write([]byte("some-bytes"))
		}))

		c = &boshio.HTTPClient{BaseURL: server.URL}
	})

	AfterEach(func() {
		server.Close()
	})

	It("should make a get request to the given path", func() {
		responseBody, err := c.Get("/some/path")
		Expect(err).NotTo(HaveOccurred())
		Expect(responseBody).To(Equal([]byte("some-bytes")))
		Expect(requestURL.Path).To(Equal("/some/path"))
	})

	Context("when the BaseURL is set and the path is absolute", func() {
		It("resolves to the path argument", func() {
			otherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("some-other-bytes"))
			}))

			respBytes, err := c.Get(otherServer.URL + "/something")
			Expect(err).NotTo(HaveOccurred())
			Expect(respBytes).To(Equal([]byte("some-other-bytes")))
		})
	})

	Context("when the base url cannot be parsed cannot be created", func() {
		It("should return an error", func() {
			c.BaseURL = "%%%"
			_, err := c.Get("/something")
			Expect(err).To(BeAssignableToTypeOf(&url.Error{}))
		})
	})

	Context("when the argument url cannot be parsed cannot be created", func() {
		It("should return an error", func() {
			_, err := c.Get("%%%")
			Expect(err).To(BeAssignableToTypeOf(&url.Error{}))
		})
	})

	Context("when the server is running TLS with a self-signed cert", func() {
		var tlsServer *httptest.Server

		BeforeEach(func() {
			tlsServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("some-bytes"))
			}))
			c.BaseURL = tlsServer.URL
		})
		AfterEach(func() {
			tlsServer.Close()
		})

		Context("when SkipTLSVerify is true", func() {
			It("should succeed", func() {
				c.SkipTLSVerify = true

				responseBody, err := c.Get("/some/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(responseBody).To(Equal([]byte("some-bytes")))
			})

		})
		Context("when SkipTLSVerify is false", func() {
			It("should return an error", func() {
				// hide tls error
				tlsServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
				c.SkipTLSVerify = false

				_, err := c.Get("/some/path")
				Expect(err).To(MatchError(HaveSuffix("x509: certificate signed by unknown authority")))
			})
		})
	})

	Context("when the server responds with a non-2xx status", func() {
		It("should return an error", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}))
			c := boshio.HTTPClient{BaseURL: testServer.URL}

			_, err := c.Get("/some/path")
			Expect(err).To(MatchError(HavePrefix("server returned status code 418")))
		})
	})

	Context("when reading the response body fails", func() {
		It("should return the error", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "12345")
				w.Write([]byte("foo"))
			}))
			c := boshio.HTTPClient{BaseURL: testServer.URL}

			_, err := c.Get("/some/path")
			Expect(err).To(MatchError("unexpected EOF"))
		})
	})
})
