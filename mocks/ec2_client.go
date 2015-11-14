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
	DescribeSubnetsCall struct {
		Receives struct {
			Input *ec2.DescribeSubnetsInput
		}
		Returns struct {
			Output *ec2.DescribeSubnetsOutput
			Error  error
		}
	}
	CreateKeyPairCall struct {
		Receives struct {
			Input *ec2.CreateKeyPairInput
		}
		Returns struct {
			Output *ec2.CreateKeyPairOutput
			Error  error
		}
	}
	DeleteKeyPairCall struct {
		Receives struct {
			Input *ec2.DeleteKeyPairInput
		}
		Returns struct {
			Output *ec2.DeleteKeyPairOutput
			Error  error
		}
	}
}

func (c *EC2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	c.DescribeImagesCall.Receives.Input = input
	return c.DescribeImagesCall.Returns.Output, c.DescribeImagesCall.Returns.Error
}

func (c *EC2Client) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	c.DescribeSubnetsCall.Receives.Input = input
	return c.DescribeSubnetsCall.Returns.Output, c.DescribeSubnetsCall.Returns.Error
}

func (c *EC2Client) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	c.CreateKeyPairCall.Receives.Input = input
	return c.CreateKeyPairCall.Returns.Output, c.CreateKeyPairCall.Returns.Error
}

func (c *EC2Client) DeleteKeyPair(input *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error) {
	c.DeleteKeyPairCall.Receives.Input = input
	return c.DeleteKeyPairCall.Returns.Output, c.DeleteKeyPairCall.Returns.Error
}
