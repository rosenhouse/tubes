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
	BOSHUser          string
	AWSRegion         string
	NATInstanceID     string
	NATElasticIP      string
	VPCID             string
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
	resources.BOSHElasticIP, ok = mapping["BOSHDirectorIP"]
	if !ok {
		return resources, errors.New("missing stack resource BOSHDirectorIP")
	}
	resources.BOSHUser, ok = mapping["BOSHDirectorUser"]
	if !ok {
		return resources, errors.New("missing stack resource BOSHDirectorUser")
	}
	resources.NATInstanceID, ok = mapping["NATInstance"]
	if !ok {
		return resources, errors.New("missing stack resource NATInstance")
	}
	resources.NATElasticIP, ok = mapping["NATEIP"]
	if !ok {
		return resources, errors.New("missing stack resource NATEIP")
	}
	resources.VPCID, ok = mapping["VPC"]
	if !ok {
		return resources, errors.New("missing stack resource VPC")
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

	return resources, nil
}
