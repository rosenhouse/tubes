package application

import (
	"io"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

type awsClient interface {
	GetLatestNATBoxAMIID() (string, error)
	UpsertStack(stackName string, template string, parameters map[string]string) error
	WaitForStack(stackName string, pundit awsclient.CloudFormationStatusPundit) error
	DeleteStack(stackName string) error
	CreateKeyPair(stackName string) (string, error)
	DeleteKeyPair(stackName string) error
}

type logger interface {
	Printf(format string, v ...interface{})
	Println(a ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(a ...interface{})
}

type configStore interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}

type Application struct {
	AWSClient    awsClient
	Logger       logger
	ResultWriter io.Writer
	ConfigStore  configStore
}
