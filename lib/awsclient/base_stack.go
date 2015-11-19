package awsclient

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type BaseStackResources struct {
	AvailabilityZone  string
	BOSHSubnetCIDR    string
	BOSHSubnetID      string
	BOSHElasticIP     string
	BOSHSecurityGroup string
	AccountID         string
	BOSHAccessKey     string
	BOSHSecretKey     string
	AWSRegion         string
}

func (c *Client) GetBaseStackResources(stackName string) (BaseStackResources, error) {
	output, err := c.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return BaseStackResources{}, err
	}

	resources := BaseStackResources{}
	mapping := map[string]string{}
	for _, resource := range output.StackResources {
		mapping[*resource.LogicalResourceId] = *resource.PhysicalResourceId
		arn, err := c.ParseARN(*resource.StackId)
		if err != nil {
			return resources, err
		}
		resources.AccountID = arn.AccountID
		resources.AWSRegion = arn.Region
	}

	var ok bool
	resources.BOSHSubnetID, ok = mapping["BOSHSubnet"]
	if !ok {
		return resources, errors.New("missing stack resource BOSHSubnet")
	}
	resources.BOSHSecurityGroup, ok = mapping["BOSHSecurityGroup"]
	if !ok {
		return resources, errors.New("missing stack resource BOSHSecurityGroup")
	}
	resources.BOSHElasticIP, ok = mapping["MicroEIP"]
	if !ok {
		return resources, errors.New("missing stack resource MicroEIP")
	}

	dsOutput, err := c.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(resources.BOSHSubnetID)},
	})
	if err != nil {
		return BaseStackResources{}, err
	}
	subnet := *dsOutput.Subnets[0]
	resources.AvailabilityZone = *subnet.AvailabilityZone
	resources.BOSHSubnetCIDR = *subnet.CidrBlock

	output2, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return BaseStackResources{}, err
	}
	for _, stackOutput := range output2.Stacks[0].Outputs {
		switch *stackOutput.OutputKey {
		case "BOSHDirectorUserAccessKey":
			resources.BOSHAccessKey = *stackOutput.OutputValue
		case "BOSHDirectorUserSecretKey":
			resources.BOSHSecretKey = *stackOutput.OutputValue
		}
	}

	return resources, nil
}
