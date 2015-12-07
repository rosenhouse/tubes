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
		"PubliclyRoutableSubnetID": Parameter{
			Type:        "String",
			Description: "ID of a publicly routable subnet, usually the BOSH subnet from the base stack",
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
		"LoadBalancerSecurityGroup": {
			Type: "AWS::EC2::SecurityGroup",
			Properties: map[string]interface{}{
				"SecurityGroupIngress": []Rule{
					{
						ToPort:     80,
						FromPort:   80,
						IpProtocol: "tcp",
						CidrIp:     "0.0.0.0/0",
					},
					{
						ToPort:     2222,
						FromPort:   2222,
						IpProtocol: "tcp",
						CidrIp:     "0.0.0.0/0",
					},
					{
						ToPort:     443,
						FromPort:   443,
						IpProtocol: "tcp",
						CidrIp:     "0.0.0.0/0",
					},
				},
				"VpcId":               Ref("VPCID"),
				"GroupDescription":    "Concourse-LoadBalancer",
				"SecurityGroupEgress": []interface{}{},
			},
		},
		"LoadBalancer": {
			Type: "AWS::ElasticLoadBalancing::LoadBalancer",
			Properties: map[string]interface{}{
				"Subnets":        []interface{}{Ref("PubliclyRoutableSubnetID")},
				"SecurityGroups": []interface{}{Ref("LoadBalancerSecurityGroup")},
				"HealthCheck": map[string]string{
					"HealthyThreshold":   "2",
					"Interval":           "30",
					"Target":             "tcp:8080",
					"Timeout":            "5",
					"UnhealthyThreshold": "10",
				},
				"Listeners": []map[string]string{
					map[string]string{
						"Protocol":         "tcp",
						"LoadBalancerPort": "80",
						"InstanceProtocol": "tcp",
						"InstancePort":     "8080",
					},
					map[string]string{
						"Protocol":         "tcp",
						"LoadBalancerPort": "2222",
						"InstanceProtocol": "tcp",
						"InstancePort":     "2222",
					},
				},
			},
		},
	},
}
