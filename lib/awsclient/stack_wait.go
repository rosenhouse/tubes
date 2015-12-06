package awsclient

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type CloudFormationStatusPundit interface {
	IsHealthy(statusString string) bool
	IsComplete(statusString string) bool
}

func (c *Client) WaitForStack(stackName string, pundit CloudFormationStatusPundit) error {
	const sleepDuration = 5 * time.Second
	elapsed := 0 * time.Second

	var status string
	var stackId string = stackName

	for {
		output, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackId),
		})
		if err != nil {
			return err
		}
		if stackId == stackName {
			stackId = *output.Stacks[0].StackId
		}

		status = *output.Stacks[0].StackStatus
		if !pundit.IsHealthy(status) {
			return fmt.Errorf("stack %q has unhealthy status %q", stackName, status)
		}
		if pundit.IsComplete(status) {
			return nil
		}

		if elapsed >= c.CloudFormationWaitTimeout {
			return fmt.Errorf("timed out waiting for stack change to complete (max %s, %s).  Check CloudFormation for details.", elapsed, status)
		}
		c.Clock.Sleep(sleepDuration)
		elapsed += sleepDuration
	}
}
