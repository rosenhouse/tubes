package integration

import "github.com/aws/aws-sdk-go/service/cloudformation"

type fakeCloudFormation struct {
	*FakeAWSBackend

	Stacks map[string]cloudformation.Stack
}

func newFakeCloudFormation(parent *FakeAWSBackend) *fakeCloudFormation {
	b := &fakeCloudFormation{
		FakeAWSBackend: parent,
	}
	b.Stacks = map[string]cloudformation.Stack{}

	return b
}
