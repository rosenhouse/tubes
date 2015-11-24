package integration

import (
	"io"
	"log"
	"reflect"
	"strings"
)

type FakeAWSBackend struct {
	Logger *log.Logger

	EC2            *fakeEC2
	CloudFormation *fakeCloudFormation
}

func NewFakeAWSBackend(logWriter io.Writer) *FakeAWSBackend {
	b := &FakeAWSBackend{}
	b.Logger = log.New(logWriter, "[ Fake AWS ] ", 0)
	b.EC2 = newFakeEC2(b)
	b.CloudFormation = newFakeCloudFormation(b)

	return b
}

func (f *FakeAWSBackend) logCall(input interface{}) {
	inputType := reflect.ValueOf(input).Type().Elem()
	pkgNameParts := strings.Split(inputType.PkgPath(), "/")
	pkgShortName := pkgNameParts[len(pkgNameParts)-1]
	actionName := strings.TrimSuffix(inputType.Name(), "Input")
	f.Logger.Printf("%s.%s", pkgShortName, actionName)
}
