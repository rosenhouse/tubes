package mocks

import "github.com/aws/aws-sdk-go/service/cloudformation"

type DescribeStacksCall struct {
	Input  *cloudformation.DescribeStacksInput
	Output *cloudformation.DescribeStacksOutput
	Error  error
}

type CloudFormationClientMultiCall struct {
	DescribeStacksCallCount int
	DescribeStacksCalls     []DescribeStacksCall
}

func NewCloudFormationClientMultiCall(callCount int) *CloudFormationClientMultiCall {
	return &CloudFormationClientMultiCall{
		DescribeStacksCalls: make([]DescribeStacksCall, callCount),
	}
}

func (c *CloudFormationClientMultiCall) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	i := c.DescribeStacksCallCount

	c.DescribeStacksCalls[i].Input = input
	out := c.DescribeStacksCalls[i].Output
	err := c.DescribeStacksCalls[i].Error

	c.DescribeStacksCallCount++
	return out, err
}

func (c *CloudFormationClientMultiCall) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	panic("not implemented")
}

func (c *CloudFormationClientMultiCall) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	panic("not implemented")
}

func (c *CloudFormationClientMultiCall) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	panic("not implemented")
}

func (c *CloudFormationClientMultiCall) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	panic("not implemented")
}

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
