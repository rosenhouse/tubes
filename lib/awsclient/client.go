package awsclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type Config struct {
	Region                    string
	AccessKey                 string
	SecretKey                 string
	CloudFormationWaitTimeout time.Duration
	EndpointOverrides         map[string]string
}

func (c *Config) getEndpoint(serviceName string) (*aws.Config, error) {
	if c.EndpointOverrides == nil {
		return &aws.Config{}, nil
	}
	endpointOverride, ok := c.EndpointOverrides[serviceName]
	if !ok || endpointOverride == "" {
		return nil, fmt.Errorf("EndpointOverrides set, but missing required service %q", serviceName)
	}
	return &aws.Config{Endpoint: aws.String(endpointOverride)}, nil
}

type ec2Client interface {
	DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
	DescribeSubnets(*ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
	CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error)
	DeleteKeyPair(*ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error)
}

type cloudformationClient interface {
	DescribeStackResources(*cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
	UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
}

type iamClient interface {
	DeleteUser(*iam.DeleteUserInput) (*iam.DeleteUserOutput, error)
	CreateAccessKey(*iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error)
	DeleteAccessKey(*iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error)
	ListAccessKeys(*iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error)
}

type clock interface {
	Sleep(time.Duration)
}

type Client struct {
	EC2                       ec2Client
	CloudFormation            cloudformationClient
	IAM                       iamClient
	Clock                     clock
	CloudFormationWaitTimeout time.Duration
}

func New(config Config) (*Client, error) {
	credentials := credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, "")
	sdkConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(config.Region),
	}

	session := session.New(sdkConfig)

	if config.CloudFormationWaitTimeout == 0 {
		return nil, fmt.Errorf("AWS config CloudFormationWaitTimeout must be a positive timeout")
	}

	ec2EndpointConfig, err := config.getEndpoint("ec2")
	if err != nil {
		return nil, err
	}
	cloudformationEndpointConfig, err := config.getEndpoint("cloudformation")
	if err != nil {
		return nil, err
	}
	iamEndpointConfig, err := config.getEndpoint("iam")
	if err != nil {
		return nil, err
	}

	return &Client{
		EC2:            ec2.New(session, ec2EndpointConfig),
		CloudFormation: cloudformation.New(session, cloudformationEndpointConfig),
		IAM:            iam.New(session, iamEndpointConfig),
		Clock:          clockImpl{},
		CloudFormationWaitTimeout: config.CloudFormationWaitTimeout,
	}, nil
}

type clockImpl struct{}

func (c clockImpl) Sleep(d time.Duration) { time.Sleep(d) }

// ARN represents an Amazon Resource Name
// http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
type ARN struct {
	Partition string
	Service   string
	Region    string
	AccountID string
	Resource  string
}

// ParseARN parses an ARN string into its component fields
func (c Client) ParseARN(arn string) (ARN, error) {
	const numExpectedParts = 6
	parts := strings.SplitN(arn, ":", numExpectedParts)
	if len(parts) < numExpectedParts {
		return ARN{}, fmt.Errorf("malformed ARN %q", arn)
	}
	return ARN{
		Partition: parts[1],
		Service:   parts[2],
		Region:    parts[3],
		AccountID: parts[4],
		Resource:  parts[5],
	}, nil
}
