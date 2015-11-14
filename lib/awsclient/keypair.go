package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *Client) CreateKeyPair(keyName string) (string, error) {
	output, err := c.EC2.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return "", err
	}

	return *output.KeyMaterial, nil
}

func (c *Client) DeleteKeyPair(keyName string) error {
	_, err := c.EC2.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(keyName),
	})
	return err
}
