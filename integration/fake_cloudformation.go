package integration

import (
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/rosenhouse/tubes/aws_enemy"
)

type FakeCloudFormation struct {
	*AWSCallLogger

	Stacks []*cloudformation.Stack
}

func NewFakeCloudFormation(logger *AWSCallLogger) *FakeCloudFormation {
	return &FakeCloudFormation{
		AWSCallLogger: logger,
	}
}

func (f *FakeCloudFormation) findStack(nameOrID string) *cloudformation.Stack {
	for _, v := range f.Stacks {
		if nameOrID == *v.StackName || nameOrID == *v.StackId {
			return v
		}
	}
	return nil
}

func (f *FakeCloudFormation) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	f.logCall(input)

	stackName := aws.StringValue(input.StackName)
	if stackName == "" {
		return &cloudformation.DescribeStacksOutput{Stacks: f.Stacks}, nil
	}

	stack := f.findStack(stackName)
	if stack != nil {
		return &cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{stack},
		}, nil
	}

	return nil, aws_enemy.CloudFormation{}.DescribeStackResources_StackMissingError(stackName)
}

func (f *FakeCloudFormation) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	f.logCall(input)

	stackName := aws.StringValue(input.StackName)
	stack := f.findStack(stackName)
	if stack != nil {
		return nil, aws_enemy.CloudFormation{}.CreateStack_AlreadyExistsError(stackName)
	}

	newStackId := aws.String(fmt.Sprintf("%x", rand.Int31()))
	newStack := &cloudformation.Stack{
		StackName:   input.StackName,
		StackId:     newStackId,
		StackStatus: aws.String("CREATE_COMPLETE"),
	}
	f.Stacks = append(f.Stacks, newStack)

	return &cloudformation.CreateStackOutput{
		StackId: newStackId,
	}, nil
}

func (f *FakeCloudFormation) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	f.logCall(input)

	stackName := aws.StringValue(input.StackName)
	stack := f.findStack(stackName)
	if stack != nil {
		stack.StackStatus = aws.String("DELETE_COMPLETE")
	}

	return &cloudformation.DeleteStackOutput{}, nil
}

func (f *FakeCloudFormation) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	f.logCall(input)

	stackName := aws.StringValue(input.StackName)
	stack := f.findStack(stackName)
	if stack == nil {
		return nil, aws_enemy.CloudFormation{}.DescribeStackResources_StackMissingError(stackName)
	}

	return &cloudformation.DescribeStackResourcesOutput{
		StackResources: []*cloudformation.StackResource{
			&cloudformation.StackResource{
				LogicalResourceId:  aws.String("BOSHSubnet"),
				PhysicalResourceId: aws.String("subnet-12345"),
				StackId:            aws.String("arn:aws:cloudformation:us-west-2:123456789012:stack/MyProductionStack/abc9dbf0-43c2-11e3-a6e8-50fa526be49c"),
			},
			&cloudformation.StackResource{
				LogicalResourceId:  aws.String("BOSHSecurityGroup"),
				PhysicalResourceId: aws.String("sg-1234"),
				StackId:            aws.String("arn:aws:cloudformation:us-west-2:123456789012:stack/MyProductionStack/abc9dbf0-43c2-11e3-a6e8-50fa526be49c"),
			},
			&cloudformation.StackResource{
				LogicalResourceId:  aws.String("MicroEIP"),
				PhysicalResourceId: aws.String("192.168.12.13"),
				StackId:            aws.String("arn:aws:cloudformation:us-west-2:123456789012:stack/MyProductionStack/abc9dbf0-43c2-11e3-a6e8-50fa526be49c"),
			},
			&cloudformation.StackResource{
				LogicalResourceId:  aws.String("BOSHDirectorUser"),
				PhysicalResourceId: aws.String("some-iam-user"),
				StackId:            aws.String("arn:aws:cloudformation:us-west-2:123456789012:stack/MyProductionStack/abc9dbf0-43c2-11e3-a6e8-50fa526be49c"),
			},
		},
	}, nil
}
