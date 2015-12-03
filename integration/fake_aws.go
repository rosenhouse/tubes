package integration

import (
	"encoding/json"
	"net/http/httptest"

	"github.com/rosenhouse/awsfaker"
)

type FakeAWS struct {
	CloudFormation *FakeCloudFormation
	EC2            *FakeEC2
	IAM            *FakeIAM

	servers map[string]*httptest.Server
}

func (f *FakeAWS) EndpointOverridesEnvVar() string {
	overrides := map[string]string{}
	for serviceName, testServer := range f.servers {
		overrides[serviceName] = testServer.URL
	}
	jsonBytes, _ := json.Marshal(overrides)
	return string(jsonBytes)
}

func NewFakeAWS(logger *AWSCallLogger) *FakeAWS {
	f := &FakeAWS{
		CloudFormation: NewFakeCloudFormation(logger),
		EC2:            NewFakeEC2(logger),
		IAM:            NewFakeIAM(logger),
	}
	f.servers = map[string]*httptest.Server{
		"cloudformation": httptest.NewServer(awsfaker.New(f.CloudFormation)),
		"ec2":            httptest.NewServer(awsfaker.New(f.EC2)),
		"iam":            httptest.NewServer(awsfaker.New(f.IAM)),
	}

	return f
}

func (f *FakeAWS) Close() {
	if f == nil || f.servers == nil {
		return
	}
	for _, server := range f.servers {
		server.Close()
	}
}
