package awsclient

import . "github.com/awslabs/aws-cfn-go-template"

var ConcourseStackTemplate = Template{
	AWSTemplateFormatVersion: "2010-09-09",
	Description:              "Infrastructure required to bootstrap a Concourse deployment, on top of an existing Base Stack for BOSH",
	Parameters: map[string]Parameter{
		"VPCID": Parameter{
			Type:        "AWS::EC2::VPC::Id",
			Description: "VPC ID to create the subnet and security group inside of",
		},
		"VPCCIDR": Parameter{
			Type:        "String",
			Default:     "10.0.0.0/16",
			Description: "CIDR block of the parent VPC",
		},
		"NATInstance": Parameter{
			Type:        "AWS::EC2::Instance::Id",
			Description: "Instance ID of NAT box",
		},
		"ConcourseSubnetCIDR": Parameter{
			Type:        "String",
			Default:     "10.0.16.0/24",
			Description: "CIDR block for the Concourse subnet",
		},
	},
	Resources: map[string]Resource{
		"ConcourseSecurityGroup": {
			Type: "AWS::EC2::SecurityGroup",
			Properties: map[string]interface{}{
				"SecurityGroupIngress": []Rule{
					{
						ToPort:     65535,
						FromPort:   0,
						IpProtocol: "-1",
						CidrIp:     Ref("VPCCIDR"),
					},
				},
				"VpcId":               Ref("VPCID"),
				"GroupDescription":    "Concourse",
				"SecurityGroupEgress": []interface{}{},
			},
		},
		"ConcourseSubnetRouteTableAssociation": {
			Type: "AWS::EC2::SubnetRouteTableAssociation",
			Properties: map[string]interface{}{
				"SubnetId":     Ref("ConcourseSubnet"),
				"RouteTableId": Ref("ConcourseRouteTable"),
			},
		},
		"ConcourseSubnet": {
			Type: "AWS::EC2::Subnet",
			Properties: map[string]interface{}{
				"VpcId":     Ref("VPCID"),
				"CidrBlock": Ref("ConcourseSubnetCIDR"),
				"Tags":      []Tag{{Key: "Name", Value: "Concourse"}},
			},
		},
		"ConcourseOutboundRoute": {
			Type: "AWS::EC2::Route",
			Properties: map[string]interface{}{
				"InstanceId":           Ref("NATInstance"),
				"DestinationCidrBlock": "0.0.0.0/0",
				"RouteTableId":         Ref("ConcourseRouteTable"),
			},
		},
		"ConcourseRouteTable": {
			Type: "AWS::EC2::RouteTable",
			Properties: map[string]interface{}{
				"VpcId": Ref("VPCID"),
			},
		},
	},
}
