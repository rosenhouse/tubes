package mocks

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2Client struct {
	DescribeImagesCall struct {
		Receives struct {
			Input *ec2.DescribeImagesInput
		}
		Returns struct {
			Output *ec2.DescribeImagesOutput
			Error  error
		}
	}
}

func (c *EC2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	c.DescribeImagesCall.Receives.Input = input
	return c.DescribeImagesCall.Returns.Output, c.DescribeImagesCall.Returns.Error
}
