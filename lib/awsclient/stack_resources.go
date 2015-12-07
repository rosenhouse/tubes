package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	StackResourceRegion    = "AWSRegion"
	StackResourceAccountID = "AccountID"
)

func (c *Client) GetStackResources(stackName string) (map[string]string, error) {
	output, err := c.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, err
	}

	resources := map[string]string{}
	for _, resource := range output.StackResources {
		resources[*resource.LogicalResourceId] = *resource.PhysicalResourceId
		arn, err := c.ParseARN(*resource.StackId)
		if err != nil {
			return nil, err
		}
		resources[StackResourceAccountID] = arn.AccountID
		resources[StackResourceRegion] = arn.Region
	}

	return resources, nil
}
