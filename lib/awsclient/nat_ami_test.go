package awsclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Discovering the latest NAT AMI ID", func() {
	var (
		client    awsclient.Client
		ec2Client *mocks.EC2Client
	)

	var makeImage = func(id, arch, volumeType, creationDate string) *ec2.Image {
		return &ec2.Image{
			Architecture: aws.String(arch),
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				&ec2.BlockDeviceMapping{
					Ebs: &ec2.EbsBlockDevice{
						VolumeType: aws.String(volumeType),
					},
				},
			},
			CreationDate: aws.String(creationDate),
			ImageId:      aws.String(id),
		}
	}

	BeforeEach(func() {
		ec2Client = &mocks.EC2Client{}
		ec2Client.DescribeImagesCall.Returns.Output = &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{
				makeImage("correct", "x86_64", "standard", "2013-10-10T22:35:35.000Z"),
				makeImage("too-old", "x86_64", "standard", "2011-10-10T22:35:35.000Z"),
				makeImage("newer-but-wrong-arch", "ARM", "standard", "2015-10-10T22:35:35.000Z"),
				makeImage("newer-but-wrong-volume-type", "x86_64", "superfast-volume", "2016-10-10T22:35:35.000Z"),
			},
		}

		client = awsclient.Client{EC2: ec2Client}
	})

	It("should return the ID of the most recent NAT AMI with the correct specs", func() {
		id, err := client.GetLatestNATBoxAMIID()
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal("correct"))
	})

	Context("when aws-sdk-go misbehaves", func() {
		Context("when the response is nil", func() {
			It("should not panic", func() {
				ec2Client.DescribeImagesCall.Returns.Output = nil
				_, err := client.GetLatestNATBoxAMIID()
				Expect(err).To(MatchError("nil response from aws-sdk-go"))
			})
		})
	})
	Context("when there are no matching images", func() {
		It("should return an error", func() {
			ec2Client.DescribeImagesCall.Returns.Output = &ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					makeImage("wrong-arch", "ARM", "standard", "2015-10-10T22:35:35.000Z"),
					makeImage("wrong-volume-type", "x86_64", "superfast-volume", "2016-10-10T22:35:35.000Z"),
				},
			}
			_, err := client.GetLatestNATBoxAMIID()
			Expect(err).To(MatchError("no AMIs found with correct specs"))
		})
	})
})
