package integration

import (
	"io"
	"log"
	"reflect"
	"strings"
)

type AWSCallLogger log.Logger

func NewAWSCallLogger(logWriter io.Writer) *AWSCallLogger {
	return (*AWSCallLogger)(log.New(logWriter, "[ Fake AWS ] ", 0))
}

func (l *AWSCallLogger) logCall(input interface{}) {
	inputType := reflect.ValueOf(input).Type().Elem()
	pkgNameParts := strings.Split(inputType.PkgPath(), "/")
	pkgShortName := pkgNameParts[len(pkgNameParts)-1]
	actionName := strings.TrimSuffix(inputType.Name(), "Input")
	(*log.Logger)(l).Printf("%s.%s", pkgShortName, actionName)
}
