package mocks

type CredentialsGenerator struct {
	FillCallback func(interface{}) error
}

func (g *CredentialsGenerator) Fill(toFill interface{}) error {
	if g.FillCallback == nil {
		panic("test setup error: missing callback")
	}
	return g.FillCallback(toFill)
}
