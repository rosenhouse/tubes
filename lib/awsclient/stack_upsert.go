package awsclient

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func formatParameters(parameters map[string]string) []*cloudformation.Parameter {
	parameterSlice := []*cloudformation.Parameter{}
	for key, value := range parameters {
		parameterSlice = append(parameterSlice, &cloudformation.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		})
	}
	return parameterSlice
}

func (c *Client) createStack(stackName string, template string, parameters map[string]string) error {
	_, err := c.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(template),
		Parameters:   formatParameters(parameters),
		Tags: []*cloudformation.Tag{
			&cloudformation.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(stackName),
			},
		},
	})
	return err
}

func errorIsBecauseNoOp(err error) bool {
	awsErr, ok := err.(awserr.RequestFailure)
	if ok && awsErr != nil {
		return awsErr.StatusCode() == 400 &&
			awsErr.Code() == "ValidationError" &&
			awsErr.Message() == "No updates are to be performed."
	}

	return false
}

func (c *Client) updateStack(stackName string, template string, parameters map[string]string) error {
	_, err := c.CloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(template),
		Parameters:   formatParameters(parameters),
	})
	if errorIsBecauseNoOp(err) {
		return nil
	}
	return err
}

func errorIsBecauseStackDoesNotExist(err error) bool {
	awsErr, ok := err.(awserr.RequestFailure)
	if !ok {
		return false
	}
	return awsErr.Code() == "ValidationError" && strings.Contains(awsErr.Message(), "does not exist")
}

func (c *Client) UpsertStack(stackName string, template string, parameters map[string]string) error {
	output, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		if errorIsBecauseStackDoesNotExist(err) {
			return c.createStack(stackName, template, parameters)
		}
		return err
	}

	status := *output.Stacks[0].StackStatus
	pundit := CloudFormationUpsertPundit{}
	if pundit.IsHealthy(status) && pundit.IsComplete(status) {
		return c.updateStack(stackName, template, parameters)
	}

	return fmt.Errorf("refusing to update stack %q, status %q", stackName, status)
}
