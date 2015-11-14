package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func (c *Client) DeleteStack(stackName string) error {
	_, err := c.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	return err
}
