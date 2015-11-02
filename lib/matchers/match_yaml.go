package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"gopkg.in/yaml.v2"
)

func MatchYAML(expected interface{}) types.GomegaMatcher {
	return &MatchYAMLMatcher{expected}
}

type MatchYAMLMatcher struct {
	YAMLToMatch interface{}
}

func (matcher *MatchYAMLMatcher) Match(actual interface{}) (success bool, err error) {
	actualString, expectedString, err := matcher.prettyPrint(actual)
	if err != nil {
		return false, err
	}

	var aval interface{}
	var eval interface{}

	// this is guarded by prettyPrint
	yaml.Unmarshal([]byte(actualString), &aval)
	yaml.Unmarshal([]byte(expectedString), &eval)

	return reflect.DeepEqual(aval, eval), nil
}

func (matcher *MatchYAMLMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return format.Message(actualString, "to match YAML of", expectedString)
}

func (matcher *MatchYAMLMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualString, expectedString, _ := matcher.prettyPrint(actual)
	return format.Message(actualString, "not to match YAML of", expectedString)
}

func (matcher *MatchYAMLMatcher) prettyPrint(actual interface{}) (actualFormatted, expectedFormatted string, err error) {
	actualString, aok := toString(actual)
	expectedString, eok := toString(matcher.YAMLToMatch)

	if !(aok && eok) {
		return "", "", fmt.Errorf("MatchYAMLMatcher matcher requires a string or stringer.  Got:\n%s", format.Object(actual, 1))
	}

	var adata interface{}
	if err := yaml.Unmarshal([]byte(actualString), &adata); err != nil {
		return "", "", err
	}
	abuf, _ := yaml.Marshal(adata)

	var edata interface{}
	if err := yaml.Unmarshal([]byte(expectedString), &edata); err != nil {
		return "", "", err
	}
	ebuf, _ := yaml.Marshal(edata)

	return string(abuf), string(ebuf), nil
}

func toString(a interface{}) (string, bool) {
	aString, isString := a.(string)
	if isString {
		return aString, true
	}

	aBytes, isBytes := a.([]byte)
	if isBytes {
		return string(aBytes), true
	}

	aStringer, isStringer := a.(fmt.Stringer)
	if isStringer {
		return aStringer.String(), true
	}

	return "", false
}
