package awsclient

import . "github.com/aws/aws-sdk-go/service/cloudformation/template"

var BaseStackTemplate = Template{
	AWSTemplateFormatVersion: "2010-09-09",
	Description:              "Infrastructure required to bootstrap a BOSH director",
	Parameters: map[string]Parameter{
		"KeyName": Parameter{
			Type:        "AWS::EC2::KeyPair::KeyName",
			Description: "Name of existing SSH Keypair to use for instances",
		},
		"NATInstanceAMI": Parameter{
			Type:        "String",
			Description: "AMI ID for NAT instance.  Use the get-nat-ami tool to find the current AMI ID",
		},
		"NATInstanceType": Parameter{
			Type:        "String",
			Default:     "t2.small",
			Description: "EC2 instance type for NAT instance",
		},
		"BOSHInboundCIDR": Parameter{
			Type:        "String",
			Default:     "0.0.0.0/0",
			Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
		},
		"VPCCIDR": Parameter{
			Type:        "String",
			Default:     "10.0.0.0/16",
			Description: "CIDR block for the VPC.",
		},
		"BOSHSubnetCIDR": Parameter{
			Type:        "String",
			Default:     "10.0.0.0/24",
			Description: "CIDR block for the BOSH subnet.",
		},
		"PrivateSubnetCIDR": Parameter{
			Type:        "String",
			Default:     "10.0.1.0/24",
			Description: "CIDR block for the private subnet.",
		},
	},
	Resources: map[string]Resource{
		"NATSecurityGroup": {
			Type: "AWS::EC2::SecurityGroup",
			Properties: map[string]interface{}{
				"SecurityGroupIngress": []Rule{
					{
						ToPort:     65535,
						FromPort:   0,
						IpProtocol: "-1",
						CidrIp:     Ref("VPCCIDR"),
					},
					{
						ToPort:     22,
						FromPort:   22,
						IpProtocol: "tcp",
						CidrIp:     Ref("BOSHInboundCIDR"),
					},
				},
				"VpcId":               Ref("VPC"),
				"GroupDescription":    "NAT",
				"SecurityGroupEgress": []interface{}{},
			},
		},
		"BOSHSecurityGroup": {
			Type: "AWS::EC2::SecurityGroup",
			Properties: map[string]interface{}{
				"SecurityGroupIngress": []Rule{
					{
						ToPort:     22,
						FromPort:   22,
						IpProtocol: "tcp",
						CidrIp:     Ref("BOSHInboundCIDR"),
					},
					{
						ToPort:     6868,
						FromPort:   6868,
						IpProtocol: "tcp",
						CidrIp:     Ref("BOSHInboundCIDR"),
					},
					{
						ToPort:     25555,
						FromPort:   25555,
						IpProtocol: "tcp",
						CidrIp:     Ref("BOSHInboundCIDR"),
					},
					{
						ToPort:     65535,
						FromPort:   0,
						IpProtocol: "-1",
						CidrIp:     Ref("VPCCIDR"),
					},
				},
				"VpcId":               Ref("VPC"),
				"GroupDescription":    "BOSH",
				"SecurityGroupEgress": []interface{}{},
			},
		},
		"NATEIP": {
			Type: "AWS::EC2::EIP",
			Properties: map[string]interface{}{
				"InstanceId": Ref("NATInstance"),
				"Domain":     "vpc",
			},
		},
		"VPC": {
			Type: "AWS::EC2::VPC",
			Properties: map[string]interface{}{
				"CidrBlock": Ref("VPCCIDR"),
			},
		},
		"PrivateSubnetRouteTableAssociation": {
			Type: "AWS::EC2::SubnetRouteTableAssociation",
			Properties: map[string]interface{}{
				"SubnetId":     Ref("PrivateSubnet"),
				"RouteTableId": Ref("PrivateRouteTable"),
			},
		},
		"PrivateSubnet": {
			Type: "AWS::EC2::Subnet",
			Properties: map[string]interface{}{
				"VpcId":     Ref("VPC"),
				"CidrBlock": Ref("PrivateSubnetCIDR"),
				"Tags":      []Tag{{Key: "Name", Value: "Private"}},
			},
		},
		"VPCGatewayAttachment": {
			Type: "AWS::EC2::VPCGatewayAttachment",
			Properties: map[string]interface{}{
				"VpcId":             Ref("VPC"),
				"InternetGatewayId": Ref("VPCGatewayInternetGateway"),
			},
		},
		"NATInstance": {
			Type: "AWS::EC2::Instance",
			Properties: map[string]interface{}{
				"SourceDestCheck":  false,
				"Tags":             []Tag{{Key: "Name", Value: "NAT"}},
				"SecurityGroupIds": []interface{}{Ref("NATSecurityGroup")},
				"KeyName":          Ref("KeyName"),
				"SubnetId":         Ref("BOSHSubnet"),
				"ImageId":          Ref("NATInstanceAMI"),
				"InstanceType":     Ref("NATInstanceType"),
			},
		},
		"VPCGatewayInternetGateway": {
			Type: "AWS::EC2::InternetGateway",
		},
		"BOSHRouteTable": {
			Type: "AWS::EC2::RouteTable",
			Properties: map[string]interface{}{
				"VpcId": Ref("VPC"),
			},
		},
		"PrivateOutboundRoute": {
			Type: "AWS::EC2::Route",
			Properties: map[string]interface{}{
				"InstanceId":           Ref("NATInstance"),
				"DestinationCidrBlock": "0.0.0.0/0",
				"RouteTableId":         Ref("PrivateRouteTable"),
			},
			DependsOn: "NATInstance",
		},
		"PrivateRouteTable": {
			Type: "AWS::EC2::RouteTable",
			Properties: map[string]interface{}{
				"VpcId": Ref("VPC"),
			},
		},
		"BOSHRoute": {
			Type: "AWS::EC2::Route",
			Properties: map[string]interface{}{
				"GatewayId":            Ref("VPCGatewayInternetGateway"),
				"RouteTableId":         Ref("BOSHRouteTable"),
				"DestinationCidrBlock": "0.0.0.0/0",
			},
			DependsOn: "VPCGatewayInternetGateway",
		},
		"BOSHSubnetRouteTableAssociation": {
			Type: "AWS::EC2::SubnetRouteTableAssociation",
			Properties: map[string]interface{}{
				"SubnetId":     Ref("BOSHSubnet"),
				"RouteTableId": Ref("BOSHRouteTable"),
			},
		},
		"BOSHSubnet": {
			Type: "AWS::EC2::Subnet",
			Properties: map[string]interface{}{
				"VpcId":     Ref("VPC"),
				"CidrBlock": Ref("BOSHSubnetCIDR"),
				"Tags":      []Tag{{Key: "Name", Value: "BOSH"}},
			},
		},
		"MicroEIP": {
			Type: "AWS::EC2::EIP",
			Properties: map[string]interface{}{
				"Domain": "vpc",
			},
		},
	},
}

type Rule struct {
	ToPort     int `json:",string"`
	FromPort   int `json:",string"`
	IpProtocol string
	CidrIp     interface{}
}

type Tag struct {
	Key   string
	Value string
}
