package awsclient_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("AWS Client", func() {
	var (
		client awsclient.Client
	)
	Describe("New()", func() {
		Context("when configured without endpoint overrides", func() {
			It("should default to the normal endpoints", func() {
				client, err := awsclient.New(awsclient.Config{
					Region: "some-region",
				})
				Expect(err).NotTo(HaveOccurred())

				ec2Client := client.EC2.(*ec2.EC2)
				Expect(ec2Client.Config.Endpoint).To(BeNil())
				cloudformationClient := client.CloudFormation.(*cloudformation.CloudFormation)
				Expect(cloudformationClient.Config.Endpoint).To(BeNil())
			})
		})
		Context("when configured with endpoint overrides", func() {
			var endpointOverrides map[string]string

			BeforeEach(func() {
				endpointOverrides = map[string]string{
					"ec2":            "http://some-fake-ec2-server.example.com:1234",
					"cloudformation": "http://some-fake-cloudformation-server.example.com:1234",
					"iam":            "http://some-fake-iam-server.example.com:1234",
				}
			})

			It("should set all the endpoints", func() {
				client, err := awsclient.New(awsclient.Config{
					Region:            "some-region",
					EndpointOverrides: endpointOverrides,
				})
				Expect(err).NotTo(HaveOccurred())

				ec2Client := client.EC2.(*ec2.EC2)
				Expect(*ec2Client.Config.Endpoint).To(Equal("http://some-fake-ec2-server.example.com:1234"))
				cloudformationClient := client.CloudFormation.(*cloudformation.CloudFormation)
				Expect(*cloudformationClient.Config.Endpoint).To(Equal("http://some-fake-cloudformation-server.example.com:1234"))
				iamClient := client.IAM.(*iam.IAM)
				Expect(*iamClient.Config.Endpoint).To(Equal("http://some-fake-iam-server.example.com:1234"))
			})
			Context("when some endpoints are missing", func() {
				It("should return an error", func() {
					endpointOverrides["cloudformation"] = ""
					client, err := awsclient.New(awsclient.Config{
						Region:            "some-region",
						EndpointOverrides: endpointOverrides,
					})
					Expect(client).To(BeNil())
					Expect(err).To(MatchError(`EndpointOverrides set, but missing required service "cloudformation"`))
				})
			})
		})

	})

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
