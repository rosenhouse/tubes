package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Config struct {
	Region           string
	AccessKey        string
	SecretKey        string
	EndpointOverride string
}

type ec2Client interface {
	DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
	DescribeSubnets(*ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
}

type cloudformationClient interface {
	DescribeStackResources(*cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
	UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
}

type Client struct {
	EC2            ec2Client
	CloudFormation cloudformationClient
}

func New(c Config) *Client {
	credentials := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	sdkConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(c.Region),
		Endpoint:    aws.String(c.EndpointOverride),
	}

	session := session.New(sdkConfig)
	ec2Client := ec2.New(session)
	cloudFormationClient := cloudformation.New(session)

	return &Client{
		EC2:            ec2Client,
		CloudFormation: cloudFormationClient,
	}
}
