package integration

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

type FakeIAM struct {
	*AWSCallLogger

	AccessKeys map[string][]string
}

func NewFakeIAM(logger *AWSCallLogger) *FakeIAM {
	return &FakeIAM{
		AWSCallLogger: logger,

		AccessKeys: map[string][]string{},
	}
}

func (f *FakeIAM) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	f.logCall(input)

	return &iam.CreateAccessKeyOutput{
		AccessKey: &iam.AccessKey{
			AccessKeyId:     aws.String("some-access-key"),
			SecretAccessKey: aws.String("some-secret-key"),
		},
	}, nil
}

func (f *FakeIAM) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	f.logCall(input)

	return &iam.DeleteAccessKeyOutput{}, nil
}

func (f *FakeIAM) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	f.logCall(input)

	return &iam.ListAccessKeysOutput{
		AccessKeyMetadata: []*iam.AccessKeyMetadata{
			&iam.AccessKeyMetadata{
				AccessKeyId: aws.String("some-iam-user-access-key"),
			},
		},
	}, nil
}
