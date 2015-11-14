package acceptance_test

import (
	"fmt"
	"math/rand"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var pathToCLI string

func getEnvironment() map[string]string {
	requiredEnvVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_DEFAULT_REGION",
	}
	values := make(map[string]string)
	missing := []string{}
	for _, name := range requiredEnvVars {
		value := os.Getenv(name)
		if value == "" {
			missing = append(missing, name)
		} else {
			values[name] = value
		}
	}
	if len(missing) > 0 {
		Fail(fmt.Sprintf("Missing required env vars for tests: %s", missing))
	}
	return values
}

var _ = BeforeSuite(func() {
	getEnvironment() // early, quick validation that all required vars are set

	rand.Seed(config.GinkgoConfig.RandomSeed)

	var err error
	pathToCLI, err = gexec.Build("github.com/rosenhouse/tubes")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
