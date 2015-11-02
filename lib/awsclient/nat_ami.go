package awsclient

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
