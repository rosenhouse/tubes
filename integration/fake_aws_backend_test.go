package integration_test

import "github.com/aws/aws-sdk-go/service/ec2"

type FakeAWSBackend struct {
	CreateKeyPairCall struct {
		Receives      *ec2.CreateKeyPairInput
		ReturnsResult *ec2.CreateKeyPairOutput
		ReturnsError  error
	}
}

func (f *FakeAWSBackend) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	f.CreateKeyPairCall.Receives = input
	return f.CreateKeyPairCall.ReturnsResult, f.CreateKeyPairCall.ReturnsError
}
