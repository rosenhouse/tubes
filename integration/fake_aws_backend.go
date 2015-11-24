package integration

import (
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/rosenhouse/awsfaker"
)

type FakeAWSBackend struct {
	Logger *log.Logger

	*awsfaker.Backend
}

func NewFakeAWSBackend(logWriter io.Writer) *FakeAWSBackend {
	b := &FakeAWSBackend{}
	b.Logger = log.New(logWriter, "[ Fake AWS ] ", 0)
	b.Backend = &awsfaker.Backend{
		EC2:            newFakeEC2(b),
		CloudFormation: newFakeCloudFormation(b),
	}

	return b
}

func (f *FakeAWSBackend) logCall(input interface{}) {
	inputType := reflect.ValueOf(input).Type().Elem()
	pkgNameParts := strings.Split(inputType.PkgPath(), "/")
	pkgShortName := pkgNameParts[len(pkgNameParts)-1]
	actionName := strings.TrimSuffix(inputType.Name(), "Input")
	f.Logger.Printf("%s.%s", pkgShortName, actionName)
}
