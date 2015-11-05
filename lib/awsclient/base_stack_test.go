package awsclient_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Retrieving resource info for the base stack", func() {
	var (
		client               awsclient.Client
		cloudFormationClient *mocks.CloudFormationClient
		ec2Client            *mocks.EC2Client
	)

	var newResource = func(logicalID, physicalID string) *cloudformation.StackResource {
		return &cloudformation.StackResource{
			LogicalResourceId:  aws.String(logicalID),
			PhysicalResourceId: aws.String(physicalID),
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
		}))
	})

	It("should use the provided stack name to look up the stack resourceS", func() {
		_, err := client.GetBaseStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(cloudFormationClient.DescribeStackResourcesCall.Receives.Input.StackName).To(Equal(aws.String("some-stack-name")))
	})

	It("should use the BOSH subnet ID from the stack to lookup details via EC2 API", func() {
		_, err := client.GetBaseStackResources("some-stack-name")
		Expect(err).NotTo(HaveOccurred())

		Expect(ec2Client.DescribeSubnetsCall.Receives.Input.SubnetIds).To(Equal([]*string{aws.String("subnet-12345")}))
	})

	Context("error cases", func() {
		Context("when describing the stack errors", func() {
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
	})
})
