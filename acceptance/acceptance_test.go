package acceptance_test

import (
	"fmt"
	"math/rand"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("The CLI", func() {
	var (
		stackName string
		envVars   map[string]string
	)

	var start = func(envVars map[string]string, args ...string) *gexec.Session {
		command := exec.Command(pathToCLI, args...)
		command.Env = []string{}
		if envVars != nil {
			for k, v := range envVars {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	BeforeEach(func() {
		stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
		envVars = getEnvironment()
	})

	It("should support basic environment manipulation", func() { // slow happy path
		const NormalTimeout = "10s"
		const StackChangeTimeout = "4m"

		By("booting a fresh environment", func() {
			session := start(envVars, "up", "bosh")

			Eventually(session.Out, NormalTimeout).Should(gbytes.Say("Looking for latest AWS NAT box AMI"))
			Eventually(session.Out, NormalTimeout).Should(gbytes.Say("ami-[a-f0-9]"))
			Eventually(session.Out, NormalTimeout).Should(gbytes.Say("Upserting stack"))
			Eventually(session.Out, StackChangeTimeout).Should(gbytes.Say("Finished"))
			Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		})
	})

	Context("invalid user input", func() { // fast failing cases
		const ErrTimeout = "10s"
		Context("no command line argument are provided", func() {
			It("should print a useful error", func() {
				session := start(nil, []string{}...)
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(ContainSubstring("usage: tubes action stack-name"))
			})
		})

		Context("when the action is unknown", func() {
			It("should print a useful error", func() {
				session := start(envVars, "nonsense_action", stackName)
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(ContainSubstring(`invalid action "nonsense_action"`))
			})
		})

		Context("when required env vars are missing", func() {
			It("should print out all the missing ones", func() {
				delete(envVars, "AWS_SECRET_ACCESS_KEY")
				envVars["AWS_DEFAULT_REGION"] = "" // two different ways to miss an env var!

				session := start(envVars, "up", stackName)

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(ContainSubstring("missing required environment variable"))
				Expect(session.Out.Contents()).To(ContainSubstring("AWS_DEFAULT_REGION"))
				Expect(session.Out.Contents()).To(ContainSubstring("AWS_SECRET_ACCESS_KEY"))
			})
		})

		Context("when the stack name is invalid", func() {
			It("should return a useful error", func() {
				session := start(envVars, "up", "invalid_stack_name")
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(ContainSubstring("invalid name: must match pattern"))
			})
		})

		Context("when application errors", func() {
			It("should inform the user", func() {
				envVars["AWS_SECRET_ACCESS_KEY"] = "some-invalid-key"

				session := start(envVars, "up", stackName)

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Out.Contents()).To(ContainSubstring("AWS was not able to validate the provided access credentials"))
			})
		})
	})
})
