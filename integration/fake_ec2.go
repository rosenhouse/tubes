package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rosenhouse/awsfaker"
	"github.com/rosenhouse/tubes/aws_enemy"
)

type FakeEC2 struct {
	*AWSCallLogger

	KeyPairs map[string]string
	Images   []*ec2.Image
}

func NewFakeEC2(logger *AWSCallLogger) *FakeEC2 {
	return &FakeEC2{
		AWSCallLogger: logger,

		KeyPairs: map[string]string{},

		Images: []*ec2.Image{
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
		},
	}
}

// generate returns a new 1024-bit RSA private key, PEM-encoded
func generate() (string, error) {
	private, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(private)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

func (f *FakeEC2) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	f.logCall(input)

	keyName := *input.KeyName
	if _, ok := f.KeyPairs[keyName]; ok {
		return nil, aws_enemy.EC2{}.CreateKeyPair_AlreadyExistsError(keyName)
	}

	var err error
	if f.KeyPairs[keyName], err = generate(); err != nil {
		return nil, &awsfaker.ErrorResponse{
			HTTPStatusCode:  500,
			AWSErrorCode:    "InternalError",
			AWSErrorMessage: err.Error(),
		}
	}

	return &ec2.CreateKeyPairOutput{
		KeyName:        input.KeyName,
		KeyFingerprint: aws.String("some-key-fingerprint"),
		KeyMaterial:    aws.String(f.KeyPairs[keyName]),
	}, nil
}

func (f *FakeEC2) DeleteKeyPair(input *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error) {
	return &ec2.DeleteKeyPairOutput{}, nil
}

func (f *FakeEC2) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	f.logCall(input)
	return &ec2.DescribeImagesOutput{
		Images: f.Images,
	}, nil
}

func (f *FakeEC2) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	f.logCall(input)

	return &ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			&ec2.Subnet{
				AvailabilityZone: aws.String("some-availability-zone"),
				CidrBlock:        aws.String("10.1.2.0/24"),
			},
		},
	}, nil
}
