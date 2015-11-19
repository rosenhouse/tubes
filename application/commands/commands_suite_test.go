package commands_test

import (
	"math/rand"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var (
	thisDir string
)

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)

	var err error
	thisDir, err = os.Getwd()
	Expect(err).NotTo(HaveOccurred())
})
