package mocks

import "github.com/aws/aws-sdk-go/service/cloudformation"

type CloudFormationClient struct {
	DescribeStackResourcesCall struct {
		Receives struct {
			Input *cloudformation.DescribeStackResourcesInput
		}
		Returns struct {
			Output *cloudformation.DescribeStackResourcesOutput
			Error  error
		}
	}
}

func (c *CloudFormationClient) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	c.DescribeStackResourcesCall.Receives.Input = input
	return c.DescribeStackResourcesCall.Returns.Output, c.DescribeStackResourcesCall.Returns.Error
}
