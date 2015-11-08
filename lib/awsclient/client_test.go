package awsclient_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("AWS Client", func() {
	var (
		client awsclient.Client
	)

	Describe("ParseARN", func() {
		It("should parse basic ARNs", func() {
			arnFormat0 := "arn:partition:service:region:account-id:resource"
			Expect(client.ParseARN(arnFormat0)).To(Equal(awsclient.ARN{
				Partition: "partition",
				Service:   "service",
				Region:    "region",
				AccountID: "account-id",
				Resource:  "resource",
			}))
		})

		It("should group the resourcetype and resource together when they are colon-separated", func() {
			arnFormat1 := "arn:partition:service:region:account-id:resourcetype:resource"
			Expect(client.ParseARN(arnFormat1)).To(Equal(awsclient.ARN{
				Partition: "partition",
				Service:   "service",
				Region:    "region",
				AccountID: "account-id",
				Resource:  "resourcetype:resource",
			}))
		})

		It("should group the resourcetype and resource together when they are slash separated", func() {
			arnFormat2 := "arn:partition:service:region:account-id:resourcetype/resource"
			Expect(client.ParseARN(arnFormat2)).To(Equal(awsclient.ARN{
				Partition: "partition",
				Service:   "service",
				Region:    "region",
				AccountID: "account-id",
				Resource:  "resourcetype/resource",
			}))
		})

		It("should handle resources with arbitrary number of slashes", func() {
			sampleARN := "arn:aws:iam::123456789012:server-certificate/division_abc/subdivision_xyz/ProdServerCert"
			Expect(client.ParseARN(sampleARN)).To(Equal(awsclient.ARN{
				Partition: "aws",
				Service:   "iam",
				Region:    "",
				AccountID: "123456789012",
				Resource:  "server-certificate/division_abc/subdivision_xyz/ProdServerCert",
			}))
		})

		Context("when the input string is malformed", func() {
			It("should return an error", func() {
				malformedARN := "arn:partition:service:region:account-id"
				_, err := client.ParseARN(malformedARN)
				Expect(err).To(MatchError(fmt.Sprintf("malformed ARN %q", malformedARN)))
			})
		})
	})
})
