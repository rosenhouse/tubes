package awsclient

const BaseStackTemplate = `
{
    "TemplateBody": {
        "AWSTemplateFormatVersion": "2010-09-09",
        "Resources": {
            "NATSecurityGroup": {
                "Type": "AWS::EC2::SecurityGroup",
                "Properties": {
                    "SecurityGroupIngress": [
                        {
                            "ToPort": "65535",
                            "FromPort": "0",
                            "IpProtocol": "-1",
                            "CidrIp": {
                                "Ref": "VPCCIDR"
                            }
                        },
                        {
                            "ToPort": "22",
                            "IpProtocol": "tcp",
                            "FromPort": "22",
                            "CidrIp": {
                                "Ref": "BOSHInboundCIDR"
                            }
                        }
                    ],
                    "VpcId": {
                        "Ref": "VPC"
                    },
                    "GroupDescription": "NAT",
                    "SecurityGroupEgress": []
                }
            },
            "BOSHSecurityGroup": {
                "Type": "AWS::EC2::SecurityGroup",
                "Properties": {
                    "SecurityGroupIngress": [
                        {
                            "ToPort": "22",
                            "FromPort": "22",
                            "IpProtocol": "tcp",
                            "CidrIp": {
                                "Ref": "BOSHInboundCIDR"
                            }
                        },
                        {
                            "ToPort": "6868",
                            "FromPort": "6868",
                            "IpProtocol": "tcp",
                            "CidrIp": {
                                "Ref": "BOSHInboundCIDR"
                            }
                        },
                        {
                            "ToPort": "25555",
                            "FromPort": "25555",
                            "IpProtocol": "tcp",
                            "CidrIp": {
                                "Ref": "BOSHInboundCIDR"
                            }
                        },
                        {
                            "ToPort": "65535",
                            "FromPort": "0",
                            "IpProtocol": "-1",
                            "CidrIp": {
                                "Ref": "VPCCIDR"
                            }
                        }
                    ],
                    "VpcId": {
                        "Ref": "VPC"
                    },
                    "GroupDescription": "BOSH",
                    "SecurityGroupEgress": []
                }
            },
            "NATEIP": {
                "Type": "AWS::EC2::EIP",
                "Properties": {
                    "InstanceId": {
                        "Ref": "NATInstance"
                    },
                    "Domain": "vpc"
                }
            },
            "VPC": {
                "Type": "AWS::EC2::VPC",
                "Properties": {
                    "CidrBlock": {
                        "Ref": "VPCCIDR"
                    }
                }
            },
            "PrivateSubnetRouteTableAssociation": {
                "Type": "AWS::EC2::SubnetRouteTableAssociation",
                "Properties": {
                    "SubnetId": {
                        "Ref": "PrivateSubnet"
                    },
                    "RouteTableId": {
                        "Ref": "PrivateRouteTable"
                    }
                }
            },
            "PrivateSubnet": {
                "Type": "AWS::EC2::Subnet",
                "Properties": {
                    "VpcId": {
                        "Ref": "VPC"
                    },
                    "CidrBlock": {
                        "Ref": "PrivateSubnetCIDR"
                    },
                    "Tags": [
                        {
                            "Value": "Private",
                            "Key": "Name"
                        }
                    ]
                }
            },
            "VPCGatewayAttachment": {
                "Type": "AWS::EC2::VPCGatewayAttachment",
                "Properties": {
                    "VpcId": {
                        "Ref": "VPC"
                    },
                    "InternetGatewayId": {
                        "Ref": "VPCGatewayInternetGateway"
                    }
                }
            },
            "NATInstance": {
                "Type": "AWS::EC2::Instance",
                "Properties": {
                    "SourceDestCheck": false,
                    "Tags": [
                        {
                            "Value": "NAT",
                            "Key": "Name"
                        }
                    ],
                    "SecurityGroupIds": [
                        {
                            "Ref": "NATSecurityGroup"
                        }
                    ],
                    "KeyName": {
                        "Ref": "KeyName"
                    },
                    "SubnetId": {
                        "Ref": "BOSHSubnet"
                    },
                    "ImageId": {
                        "Ref": "NATInstanceAMI"
                    },
                    "InstanceType": {
                        "Ref": "NATInstanceType"
                    }
                }
            },
            "VPCGatewayInternetGateway": {
                "Type": "AWS::EC2::InternetGateway"
            },
            "BOSHRouteTable": {
                "Type": "AWS::EC2::RouteTable",
                "Properties": {
                    "VpcId": {
                        "Ref": "VPC"
                    }
                }
            },
            "PrivateOutboundRoute": {
                "Type": "AWS::EC2::Route",
                "Properties": {
                    "InstanceId": {
                        "Ref": "NATInstance"
                    },
                    "DestinationCidrBlock": "0.0.0.0/0",
                    "RouteTableId": {
                        "Ref": "PrivateRouteTable"
                    }
                },
                "DependsOn": "NATInstance"
            },
            "PrivateRouteTable": {
                "Type": "AWS::EC2::RouteTable",
                "Properties": {
                    "VpcId": {
                        "Ref": "VPC"
                    }
                }
            },
            "BOSHRoute": {
                "Type": "AWS::EC2::Route",
                "Properties": {
                    "GatewayId": {
                        "Ref": "VPCGatewayInternetGateway"
                    },
                    "DestinationCidrBlock": "0.0.0.0/0",
                    "RouteTableId": {
                        "Ref": "BOSHRouteTable"
                    }
                },
                "DependsOn": "VPCGatewayInternetGateway"
            },
            "BOSHSubnetRouteTableAssociation": {
                "Type": "AWS::EC2::SubnetRouteTableAssociation",
                "Properties": {
                    "SubnetId": {
                        "Ref": "BOSHSubnet"
                    },
                    "RouteTableId": {
                        "Ref": "BOSHRouteTable"
                    }
                }
            },
            "BOSHSubnet": {
                "Type": "AWS::EC2::Subnet",
                "Properties": {
                    "VpcId": {
                        "Ref": "VPC"
                    },
                    "CidrBlock": {
                        "Ref": "BOSHSubnetCIDR"
                    },
                    "Tags": [
                        {
                            "Value": "BOSH",
                            "Key": "Name"
                        }
                    ]
                }
            },
            "MicroEIP": {
                "Type": "AWS::EC2::EIP",
                "Properties": {
                    "Domain": "vpc"
                }
            }
        },
        "Description": "Infrastructure required to bootstrap a BOSH director",
        "Parameters": {
            "BOSHSubnetCIDR": {
                "Default": "10.0.0.0/24",
                "Type": "String",
                "Description": "CIDR block for the BOSH subnet."
            },
            "BOSHInboundCIDR": {
                "Default": "0.0.0.0/0",
                "Type": "String",
                "Description": "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)"
            },
            "VPCCIDR": {
                "Default": "10.0.0.0/16",
                "Type": "String",
                "Description": "CIDR block for the VPC."
            },
            "NATInstanceType": {
                "Default": "t2.small",
                "Type": "String",
                "Description": "CIDR block for the BOSH subnet."
            },
            "KeyName": {
                "Default": "bosh",
                "Type": "AWS::EC2::KeyPair::KeyName",
                "Description": "Name of existing SSH Keypair to use for instances"
            },
            "NATInstanceAMI": {
                "Type": "String",
                "Description": "AMI ID for NAT instance.  Use the get-nat-ami tool to find the current AMI ID"
            },
            "PrivateSubnetCIDR": {
                "Default": "10.0.1.0/24",
                "Type": "String",
                "Description": "CIDR block for the private subnet."
            }
        }
    }
}
`
