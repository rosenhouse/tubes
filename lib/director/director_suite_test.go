package director_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDirector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Director Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})
