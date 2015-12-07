package awsclient_test

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Retrieving resource IDs from a CloudFormation stack", func() {
	var (
		client               awsclient.Client
		cloudFormationClient *mocks.CloudFormationClient
	)
	const (
		region              = "some-region"
		accountID    uint64 = 123456789012 // 12 digits
		stackName           = "some-stack-name"
		resourceGUID        = "some-resource-guid"
	)

	var newResource = func(logicalID, physicalID string) *cloudformation.StackResource {
		return &cloudformation.StackResource{
			LogicalResourceId:  aws.String(logicalID),
			PhysicalResourceId: aws.String(physicalID),
			StackId: aws.String(fmt.Sprintf("arn:aws:cloudformation:%s:%d:stack/%s/%s",
				region,
				accountID,
				stackName,
				resourceGUID,
			)),
		}
	}

	BeforeEach(func() {
		cloudFormationClient = &mocks.CloudFormationClient{}
		client = awsclient.Client{
			CloudFormation: cloudFormationClient,
		}

		cloudFormationClient.DescribeStackResourcesCall.Returns.Output = &cloudformation.DescribeStackResourcesOutput{
			StackResources: []*cloudformation.StackResource{
				newResource("SomeLogicalResourceID", "some-physical-resource-id"),
				newResource("SomeOtherLogicalResourceID", "some-other-physical-resource-id"),
			},
		}
	})

	It("should use the provided stack name to look up the stack resources", func() {
		_, err := client.GetStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(cloudFormationClient.DescribeStackResourcesCall.Receives.Input.StackName).To(Equal(aws.String("some-stack-name")))
	})

	It("should return all the resources", func() {
		resources, err := client.GetStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(resources).To(HaveKeyWithValue("SomeLogicalResourceID", "some-physical-resource-id"))
		Expect(resources).To(HaveKeyWithValue("SomeOtherLogicalResourceID", "some-other-physical-resource-id"))
	})

	It("should return pseudo-resources for account ID and region", func() {
		resources, err := client.GetStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(resources).To(HaveKeyWithValue("AWSRegion", "some-region"))
		Expect(resources).To(HaveKeyWithValue("AccountID", "123456789012"))
	})

	Context("error cases", func() {
		Context("when describing the stack resources errors", func() {
			It("should return the error", func() {
				cloudFormationClient.DescribeStackResourcesCall.Returns.Error = errors.New("some error")
				_, err := client.GetStackResources("some-stack-name")
				Expect(err).To(MatchError("some error"))
			})
		})

		Context("when a stack resource's StackID is not a valid ARN", func() {
			It("should return an error", func() {
				cloudFormationClient.DescribeStackResourcesCall.Returns.Output.StackResources[1].StackId = aws.String("invalid-stackid")
				_, err := client.GetStackResources("some-stack-name")
				Expect(err).To(MatchError(`malformed ARN "invalid-stackid"`))
			})
		})
	})
})
