package mocks

type CloudConfigGenerator struct {
	GenerateCall struct {
		Receives struct {
			Resources map[string]string
		}
		Returns struct {
			Bytes []byte
			Error error
		}
	}
}

func (g *CloudConfigGenerator) Generate(resources map[string]string) ([]byte, error) {
	g.GenerateCall.Receives.Resources = resources
	return g.GenerateCall.Returns.Bytes, g.GenerateCall.Returns.Error
}
