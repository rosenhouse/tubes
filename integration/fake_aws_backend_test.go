package integration_test

import (
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type FakeAWSBackend struct {
	Logger *log.Logger

	CreateKeyPairCall struct {
		Receives      *ec2.CreateKeyPairInput
		ReturnsResult *ec2.CreateKeyPairOutput
		ReturnsError  error
	}
}

func NewFakeAWSBackend(logWriter io.Writer) *FakeAWSBackend {
	b := &FakeAWSBackend{}
	b.Logger = log.New(logWriter, "[ Fake AWS ] ", 0)

	b.CreateKeyPairCall.ReturnsResult = &ec2.CreateKeyPairOutput{
		KeyName:        aws.String("some-key-name"),
		KeyFingerprint: aws.String("some-key-fingerprint"),
		KeyMaterial:    aws.String("some-pem-data"),
	}

	return b
}

func (f *FakeAWSBackend) logCall(input, output, err interface{}) {
	inputType := reflect.ValueOf(input).Type().Elem()
	pkgNameParts := strings.Split(inputType.PkgPath(), "/")
	pkgShortName := pkgNameParts[len(pkgNameParts)-1]
	actionName := strings.TrimSuffix(inputType.Name(), "Input")
	f.Logger.Printf("%s.%s", pkgShortName, actionName)
}

func (f *FakeAWSBackend) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	f.CreateKeyPairCall.Receives = input
	output, err := f.CreateKeyPairCall.ReturnsResult, f.CreateKeyPairCall.ReturnsError
	f.logCall(input, output, err)
	return output, err
}
