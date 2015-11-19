package awsclient_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Generating the base template", func() {
	It("should match the fixture", func() {
		asJSON := awsclient.BaseStackTemplate.String()

		expected, err := ioutil.ReadFile("fixtures/base_stack_template.json")
		Expect(err).NotTo(HaveOccurred())

		Expect(asJSON).To(MatchJSON(expected))
	})
})

var _ = Describe("Retrieving resource info for the base stack", func() {
	var (
		client               awsclient.Client
		cloudFormationClient *mocks.CloudFormationClient
		ec2Client            *mocks.EC2Client
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

	var newOutput = func(key, value string) *cloudformation.Output {
		return &cloudformation.Output{
			OutputKey:   aws.String(key),
			OutputValue: aws.String(value),
		}
	}

	BeforeEach(func() {
		cloudFormationClient = &mocks.CloudFormationClient{}
		ec2Client = &mocks.EC2Client{}
		client = awsclient.Client{
			EC2:            ec2Client,
			CloudFormation: cloudFormationClient,
		}

		cloudFormationClient.DescribeStackResourcesCall.Returns.Output = &cloudformation.DescribeStackResourcesOutput{
			StackResources: []*cloudformation.StackResource{
				newResource("BOSHSecurityGroup", "sg-12345"),
				newResource("BOSHSubnet", "subnet-12345"),
				newResource("MicroEIP", "54.123.456.78"),
				newResource("NATEIP", "nat-eip-ignore-this"),
				newResource("NATSecurityGroup", "nat-security-group-ignore-this"),
				newResource("PrivateSubnet", "private-subnet-ignore-this-for-now"),
			},
		}

		cloudFormationClient.DescribeStacksCall.Returns.Output = &cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					Outputs: []*cloudformation.Output{
						newOutput("BOSHDirectorUserAccessKey", "some-access-key-id"),
						newOutput("BOSHDirectorUserSecretKey", "some-secret-access-key"),
					},
				},
			},
		}

		ec2Client.DescribeSubnetsCall.Returns.Output = &ec2.DescribeSubnetsOutput{
			Subnets: []*ec2.Subnet{
				&ec2.Subnet{
					AvailabilityZone: aws.String("some-nat-az"),
					CidrBlock:        aws.String("10.11.12.13/24"),
				},
			},
		}
	})

	It("should return the AZ, BOSH Subnet, Elastic IP and Security Group", func() {
		baseStack, err := client.GetBaseStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(baseStack).To(Equal(awsclient.BaseStackResources{
			AvailabilityZone:  "some-nat-az",
			BOSHSubnetCIDR:    "10.11.12.13/24",
			BOSHSubnetID:      "subnet-12345",
			BOSHElasticIP:     "54.123.456.78",
			BOSHSecurityGroup: "sg-12345",
			AccountID:         "123456789012",
			BOSHAccessKey:     "some-access-key-id",
			BOSHSecretKey:     "some-secret-access-key",
			AWSRegion:         "some-region",
		}))
	})

	It("should use the provided stack name to look up the stack resources", func() {
		_, err := client.GetBaseStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(cloudFormationClient.DescribeStackResourcesCall.Receives.Input.StackName).To(Equal(aws.String("some-stack-name")))
		Expect(cloudFormationClient.DescribeStacksCall.Receives.Input.StackName).To(Equal(aws.String("some-stack-name")))
	})

	It("should use the BOSH subnet ID from the stack to lookup details via EC2 API", func() {
		_, err := client.GetBaseStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(ec2Client.DescribeSubnetsCall.Receives.Input.SubnetIds).To(Equal([]*string{aws.String("subnet-12345")}))
	})

	Context("error cases", func() {
		Context("when describing the stack resources errors", func() {
			It("should return the error", func() {
				cloudFormationClient.DescribeStackResourcesCall.Returns.Error = errors.New("some error")
				_, err := client.GetBaseStackResources("some-stack-name")
				Expect(err).To(MatchError("some error"))
			})
		})
		Context("when describing the subnet errors", func() {
			It("should return the error", func() {
				ec2Client.DescribeSubnetsCall.Returns.Error = errors.New("some error")
				_, err := client.GetBaseStackResources("some-stack-name")
				Expect(err).To(MatchError("some error"))
			})
		})
		Context("when expected resources are missing from the stack", func() {
			It("should return an error", func() {
				cloudFormationClient.DescribeStackResourcesCall.Returns.Output.StackResources[0] = newResource("nope", "very")
				_, err := client.GetBaseStackResources("some-stack-name")
				Expect(err).To(MatchError("missing stack resource BOSHSecurityGroup"))
			})
		})
		Context("when a stack resource's StackID is not a valid ARN", func() {
			It("should return an error", func() {
				cloudFormationClient.DescribeStackResourcesCall.Returns.Output.StackResources[3].StackId = aws.String("invalid-stackid")
				_, err := client.GetBaseStackResources("some-stack-name")
				Expect(err).To(MatchError(`malformed ARN "invalid-stackid"`))
			})
		})
		Context("when describing the stack errors", func() {
			It("should return the error", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = errors.New("some error")
				_, err := client.GetBaseStackResources("some-stack-name")
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
