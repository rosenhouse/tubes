package awsclient

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Endpoints struct {
	EC2 string
}

type Config struct {
	Region           string
	AccessKey        string
	SecretKey        string
	EndpointOverride string
}

type ec2Client interface {
	DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
}

type Client struct {
	EC2 ec2Client
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

	return &Client{
		EC2: ec2Client,
	}
}

func (c *Client) GetLatestNATBoxAMIID() (string, error) {
	resp, err := c.EC2.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("name"),
				Values: []*string{aws.String("amzn-ami-vpc-nat-hvm*")},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("couldn't describe images: %s", err)
	}

	if resp == nil {
		return "", errors.New("nil response from aws-sdk-go")
	}

	var latestImage *ec2.Image
	for _, image := range resp.Images {
		if *image.Architecture != "x86_64" {
			continue
		}
		if *image.BlockDeviceMappings[0].Ebs.VolumeType != "standard" {
			continue
		}
		if latestImage == nil || *image.CreationDate > *latestImage.CreationDate {
			latestImage = image
		}
	}

	if latestImage == nil {
		return "", errors.New("no AMIs found with correct specs")
	}

	return *latestImage.ImageId, nil
}
