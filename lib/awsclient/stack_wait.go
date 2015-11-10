package awsclient

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func (c *Client) WaitForStack(stackName string) error {
	const sleepDuration = 5 * time.Second
	elapsed := 0 * time.Second

	var status string

	for {
		output, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return err
		}
		status = *output.Stacks[0].StackStatus
		if !c.CloudFormationStatusPundit.IsHealthy(status) {
			return fmt.Errorf("stack %q has unhealthy status %q", stackName, status)
		}
		if c.CloudFormationStatusPundit.IsComplete(status) {
			return nil
		}

		if elapsed >= c.CloudFormationWaitTimeout {
			return fmt.Errorf("timed out waiting for stack change to complete (max %s, %s)", elapsed, status)
		}
		c.Clock.Sleep(sleepDuration)
		elapsed += sleepDuration
	}
}
