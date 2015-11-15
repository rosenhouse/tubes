package mocks

import "github.com/rosenhouse/tubes/lib/director"

func NewBoshIOClient(nReleaseCalls int) *BoshIOClient {
	return &BoshIOClient{LatestReleaseCalls: make([]LatestReleaseCall, nReleaseCalls)}
}

type LatestReleaseCall struct {
	Receives struct {
		ReleasePath string
	}
	Returns struct {
		Artifact director.Artifact
		Error    error
	}
}

type BoshIOClient struct {
	LatestStemcellCall struct {
		Receives struct {
			StemcellName string
		}
		Returns struct {
			Artifact director.Artifact
			Error    error
		}
	}

	LatestReleaseCalls     []LatestReleaseCall
	LatestReleaseCallCount int
}

func (c *BoshIOClient) LatestRelease(releasePath string) (director.Artifact, error) {
	i := c.LatestReleaseCallCount
	c.LatestReleaseCalls[i].Receives.ReleasePath = releasePath
	a, e := c.LatestReleaseCalls[i].Returns.Artifact, c.LatestReleaseCalls[i].Returns.Error
	c.LatestReleaseCallCount++
	return a, e
}

func (c *BoshIOClient) LatestStemcell(stemcellName string) (director.Artifact, error) {
	c.LatestStemcellCall.Receives.StemcellName = stemcellName
	return c.LatestStemcellCall.Returns.Artifact, c.LatestStemcellCall.Returns.Error
}
