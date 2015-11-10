package application_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"testing"
)

func TestApplication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})
