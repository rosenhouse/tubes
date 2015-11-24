package integration

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rosenhouse/tubes/aws_enemy"
)

type fakeEC2 struct {
	*FakeAWSBackend

	KeyPairs map[string]string
	Images   []*ec2.Image
}

func newFakeEC2(parent *FakeAWSBackend) *fakeEC2 {
	b := &fakeEC2{
		FakeAWSBackend: parent,
	}

	b.KeyPairs = map[string]string{
		"some-existing-name": "some-existing-pem-data",
	}

	b.Images = []*ec2.Image{
		&ec2.Image{
			Architecture: aws.String("x86_64"),
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				&ec2.BlockDeviceMapping{
					Ebs: &ec2.EbsBlockDevice{
						VolumeType: aws.String("standard"),
					},
				},
			},
			CreationDate: aws.String("2013-10-10T22:35:35.000Z"),
			ImageId:      aws.String("ami-whatever"),
		},
	}

	return b
}

func (f *fakeEC2) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	f.logCall(input)

	keyName := *input.KeyName
	if _, ok := f.KeyPairs[keyName]; ok {
		return nil, aws_enemy.EC2{}.CreateKeyPair_AlreadyExistsError(keyName)
	}

	f.KeyPairs[keyName] = "some-new-pem-data"

	return &ec2.CreateKeyPairOutput{
		KeyName:        input.KeyName,
		KeyFingerprint: aws.String("some-key-fingerprint"),
		KeyMaterial:    aws.String("some-pem-data"),
	}, nil
}

func (f *fakeEC2) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	f.logCall(input)
	return &ec2.DescribeImagesOutput{
		Images: f.Images,
	}, nil
}
