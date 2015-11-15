package mocks

import "github.com/rosenhouse/tubes/lib/awsclient"

type ManifestBuilder struct {
	BuildCall struct {
		Receives struct {
			StackName string
			Resources awsclient.BaseStackResources
		}
		Returns struct {
			ManifestYAML []byte
			Error        error
		}
	}
}

func (b *ManifestBuilder) Build(stackName string, resources awsclient.BaseStackResources) ([]byte, error) {
	b.BuildCall.Receives.StackName = stackName
	b.BuildCall.Receives.Resources = resources
	return b.BuildCall.Returns.ManifestYAML, b.BuildCall.Returns.Error
}
