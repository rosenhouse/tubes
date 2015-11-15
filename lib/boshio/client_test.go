package boshio_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/boshio"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("getting artifact info from the bosh.io API", func() {
	var (
		client     *boshio.Client
		jsonClient *mocks.JSONClient
		httpClient *mocks.HTTPClient
	)

	BeforeEach(func() {
		jsonClient = &mocks.JSONClient{}
		httpClient = &mocks.HTTPClient{}
		client = &boshio.Client{
			JSONClient: jsonClient,
			HTTPClient: httpClient,
		}
	})

	Describe("getting the url and sha of the latest release", func() {
		BeforeEach(func() {
			jsonClient.GetCall.ResponseJSON = `[
			{
				"name": "some-release-path",
				"version": "224",
				"url": "some-download-link"
			},
			{
				"name": "some-release-path",
				"version": "223",
				"url": "some-old-download-link"
			}
		]`

			httpClient.GetCall.Returns.Body = []byte(`... some HTML before it .... <pre>
releases:
- name: some-release-name
  url: some-download-link
	sha1: f572d396fae9206628714fb2ce00f72e94f2258f
	</pre> ... some more HTML after it.... `)
		})

		It("should return the url for downloading the latest release", func() {
			artifact, err := client.LatestRelease("some-release-path")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.GetCall.Receives.Path).To(Equal("/api/v1/releases/some-release-path"))
			Expect(artifact.URL).To(Equal("some-download-link"))
		})

		It("should scrape the SHA1 from the HTML page for the release", func() {
			artifact, err := client.LatestRelease("some-release-path")
			Expect(err).NotTo(HaveOccurred())

			Expect(artifact.SHA).To(Equal("f572d396fae9206628714fb2ce00f72e94f2258f"))
		})

		Context("when the json API errors", func() {
			It("should return the error", func() {
				jsonClient.GetCall.Returns.Error = errors.New("some error")
				_, err := client.LatestRelease("some-release-path")
				Expect(err).To(MatchError("some error"))
			})
		})

		Context("when the list of releases is empty", func() {
			It("should return an error", func() {
				jsonClient.GetCall.ResponseJSON = `[]`
				_, err := client.LatestRelease("some-release-path")
				Expect(err).To(MatchError("empty result for /api/v1/releases/some-release-path"))
			})
		})

		Context("when downloading release HTML page fails", func() {
			It("should return the error", func() {
				httpClient.GetCall.Returns.Error = errors.New("dial tcp or something")
				_, err := client.LatestRelease("some-release-path")
				Expect(err).To(MatchError("dial tcp or something"))
			})
		})

		Context("if the sha1 sum cannot be found in the release HTML page", func() {
			It("should return an informative error", func() {
				httpClient.GetCall.Returns.Body = []byte("sha1: invalid-sha")
				_, err := client.LatestRelease("some-release-path")
				Expect(err).To(MatchError(fmt.Sprintf(
					"failed while scraping %s: unable to find sha1",
					"/releases/some-release-path?version=224")))
			})
		})
	})

	Describe("getting the latest stemcell specs", func() {
		BeforeEach(func() {
			jsonClient.GetCall.ResponseJSON = `[
				{
					"light": {
						"md5": "latest-md5",
						"size": 12345,
						"url": "some-download-url"
					},
					"light_china": {
						"md5": "latest-china-md5",
						"url": "some-other-download-url"
					},
					"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
					"version": "9"
				},
				{
					"light": {
						"md5": "some-old-md5",
						"url": "some-old-download-url"
					},
					"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
					"version": "8"
				}
			]`

			httpClient.GetCall.Returns.Body = []byte("some tarball bytes")
		})

		It("should follow the download link to get the sha", func() {
			artifact, err := client.LatestStemcell("some-stemcell-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.GetCall.Receives.Path).To(Equal("/api/v1/stemcells/some-stemcell-name"))
			Expect(artifact.SHA).To(Equal("fc4fece5aacd3b1e2d170b5fdeee435749b90417"))
			Expect(artifact.URL).To(Equal("some-download-url"))
		})

		Context("when the json client fails", func() {
			It("should return the error", func() {
				jsonClient.GetCall.Returns.Error = errors.New("some error")

				_, err := client.LatestStemcell("some-stemcell-name")
				Expect(err).To(MatchError(err))
			})
		})

		Context("when the response data is empty", func() {
			It("should return an error", func() {
				jsonClient.GetCall.ResponseJSON = `[]`
				_, err := client.LatestStemcell("some-stemcell-name")
				Expect(err).To(MatchError("empty result for /api/v1/stemcells/some-stemcell-name"))
			})
		})

		Context("when downloading the stemcell fails", func() {
			It("should return the error", func() {
				httpClient.GetCall.Returns.Error = errors.New("dial tcp or something")
				_, err := client.LatestStemcell("some-stemcell-name")
				Expect(err).To(MatchError("error while downloading stemcell: dial tcp or something"))
			})
		})
	})
})
