package mocks

import "github.com/rosenhouse/tubes/lib/awsclient"

type ManifestBuilder struct {
	BuildCall struct {
		Receives struct {
			StackName string
			Resources awsclient.BaseStackResources
			AccessKey string
			SecretKey string
		}
		Returns struct {
			ManifestYAML []byte
			Error        error
		}
	}
}

func (b *ManifestBuilder) Build(stackName string, resources awsclient.BaseStackResources, accessKey, secretKey string) ([]byte, error) {
	b.BuildCall.Receives.StackName = stackName
	b.BuildCall.Receives.Resources = resources
	b.BuildCall.Receives.AccessKey = accessKey
	b.BuildCall.Receives.SecretKey = secretKey
	return b.BuildCall.Returns.ManifestYAML, b.BuildCall.Returns.Error
}
