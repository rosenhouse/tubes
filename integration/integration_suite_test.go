package integration_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/onsi/ginkgo/config"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var pathToCLI string

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)

	var err error
	pathToCLI, err = gexec.Build("github.com/rosenhouse/tubes")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
