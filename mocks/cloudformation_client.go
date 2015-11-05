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

	DescribeStacksCall struct {
		Receives struct {
			Input *cloudformation.DescribeStacksInput
		}
		Returns struct {
			Output *cloudformation.DescribeStacksOutput
			Error  error
		}
	}

	CreateStackCall struct {
		Receives struct {
			Input *cloudformation.CreateStackInput
		}
		Returns struct {
			Output *cloudformation.CreateStackOutput
			Error  error
		}
	}

	UpdateStackCall struct {
		Receives struct {
			Input *cloudformation.UpdateStackInput
		}
		Returns struct {
			Output *cloudformation.UpdateStackOutput
			Error  error
		}
	}

	DeleteStackCall struct {
		Receives struct {
			Input *cloudformation.DeleteStackInput
		}
		Returns struct {
			Output *cloudformation.DeleteStackOutput
			Error  error
		}
	}
}

func (c *CloudFormationClient) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	c.DescribeStackResourcesCall.Receives.Input = input
	return c.DescribeStackResourcesCall.Returns.Output, c.DescribeStackResourcesCall.Returns.Error
}

func (c *CloudFormationClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	c.DescribeStacksCall.Receives.Input = input
	return c.DescribeStacksCall.Returns.Output, c.DescribeStacksCall.Returns.Error
}
func (c *CloudFormationClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	c.CreateStackCall.Receives.Input = input
	return c.CreateStackCall.Returns.Output, c.CreateStackCall.Returns.Error
}
func (c *CloudFormationClient) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	c.UpdateStackCall.Receives.Input = input
	return c.UpdateStackCall.Returns.Output, c.UpdateStackCall.Returns.Error
}
func (c *CloudFormationClient) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	c.DeleteStackCall.Receives.Input = input
	return c.DeleteStackCall.Returns.Output, c.DeleteStackCall.Returns.Error
}
