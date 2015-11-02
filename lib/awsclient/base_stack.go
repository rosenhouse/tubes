package awsclient

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type BaseStack struct {
	AvailabilityZone  string
	BOSHSubnetCIDR    string
	BOSHSubnetID      string
	BOSHElasticIP     string
	BOSHSecurityGroup string
}

func (c *Client) GetBaseStackResources(stackName string) (BaseStack, error) {
	output, err := c.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String("foo"),
	})
	if err != nil {
		return BaseStack{}, err
	}

	mapping := map[string]string{}
	for _, resource := range output.StackResources {
		mapping[*resource.LogicalResourceId] = *resource.PhysicalResourceId
	}

	baseStack := BaseStack{}
	var ok bool
	baseStack.BOSHSubnetID, ok = mapping["BOSHSubnet"]
	if !ok {
		return baseStack, errors.New("missing stack resource BOSHSubnet")
	}
	baseStack.BOSHSecurityGroup, ok = mapping["BOSHSecurityGroup"]
	if !ok {
		return baseStack, errors.New("missing stack resource BOSHSecurityGroup")
	}
	baseStack.BOSHElasticIP, ok = mapping["MicroEIP"]
	if !ok {
		return baseStack, errors.New("missing stack resource MicroEIP")
	}

	dsOutput, err := c.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(baseStack.BOSHSubnetID)},
	})
	if err != nil {
		return BaseStack{}, err
	}
	subnet := *dsOutput.Subnets[0]
	baseStack.AvailabilityZone = *subnet.AvailabilityZone
	baseStack.BOSHSubnetCIDR = *subnet.CidrBlock

	return baseStack, nil
}
