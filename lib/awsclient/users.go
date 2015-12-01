package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func (c *Client) CreateAccessKey(userName string) (string, string, error) {
	output, err := c.IAM.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return "", "", err
	}
	return *output.AccessKey.AccessKeyId, *output.AccessKey.SecretAccessKey, nil
}

func (c *Client) DeleteAccessKey(userName, accessKeyID string) error {
	_, err := c.IAM.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    aws.String(userName),
		AccessKeyId: aws.String(accessKeyID),
	})

	return err
}

func (c *Client) ListAccessKeys(userName string) ([]string, error) {
	output, err := c.IAM.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, md := range output.AccessKeyMetadata {
		keys = append(keys, *md.AccessKeyId)
	}
	return keys, nil
}
