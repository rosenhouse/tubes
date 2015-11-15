package mocks

import (
	"github.com/rosenhouse/tubes/lib/director"
	"github.com/rosenhouse/tubes/lib/manifests"
)

type DirectorManifestGenerator struct {
	GenerateCall struct {
		Receives struct {
			Config director.DirectorConfig
		}
		Returns struct {
			Manifest manifests.Manifest
			Error    error
		}
	}
}

func (g *DirectorManifestGenerator) Generate(config director.DirectorConfig) (manifests.Manifest, error) {
	g.GenerateCall.Receives.Config = config
	return g.GenerateCall.Returns.Manifest, g.GenerateCall.Returns.Error
}
