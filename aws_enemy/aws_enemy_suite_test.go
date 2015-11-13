package aws_enemy_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Enemy Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})
