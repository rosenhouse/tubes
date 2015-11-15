package acceptance_test

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("The CLI", func() {
	var (
		stackName  string
		envVars    map[string]string
		workingDir string
	)

	var start = func(envVars map[string]string, args ...string) *gexec.Session {
		command := exec.Command(pathToCLI, args...)
		command.Env = []string{}
		if envVars != nil {
			for k, v := range envVars {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
		command.Dir = workingDir
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	BeforeEach(func() {
		stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
		var err error
		workingDir, err = ioutil.TempDir("", "tubes-acceptance-test")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("happy path", func() {
		BeforeEach(func() {
			stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
			envVars = getEnvironment()
		})
		AfterEach(func() {
			fmt.Fprintf(GinkgoWriter, "\n\n  -->  Cleaning up.  You may not want this if you're debugging a test failure.\n\n")
			client := awsclient.New(awsclient.Config{
				Region:    envVars["AWS_DEFAULT_REGION"],
				AccessKey: envVars["AWS_ACCESS_KEY_ID"],
				SecretKey: envVars["AWS_SECRET_ACCESS_KEY"],
			})
			Expect(client.DeleteStack(stackName)).To(Succeed())
			Expect(client.DeleteKeyPair(stackName)).To(Succeed())
		})

		It("should support basic environment manipulation", func() { // slow happy path
			const NormalTimeout = "10s"
			const StackChangeTimeout = "4m"

			By("booting a fresh environment", func() {
				session := start(envVars, "-n", stackName, "up")

				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Creating keypair"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Looking for latest AWS NAT box AMI"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("ami-[a-f0-9]*"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Finished"))
				Eventually(session, NormalTimeout).Should(gexec.Exit(0))
			})

			By("storing the SSH key on the filesystem", func() {
				Expect(ioutil.ReadFile(filepath.Join(workingDir, stackName, "ssh-key"))).To(ContainSubstring("RSA PRIVATE KEY"))
			})

			By("exposing the SSH key", func() {
				session := start(envVars, "-n", stackName, "show")

				Eventually(session, NormalTimeout).Should(gexec.Exit(0))

				pemBlock, _ := pem.Decode(session.Out.Contents())
				Expect(pemBlock).NotTo(BeNil())
				Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

				_, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
				Expect(err).NotTo(HaveOccurred())
			})

			By("tearing down the environment", func() {
				session := start(envVars, "-n", stackName, "down")

				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Delete complete"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting keypair"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
				Eventually(session, NormalTimeout).Should(gexec.Exit(0))
			})
		})
	})

	Context("invalid user input", func() { // fast failing cases
		BeforeEach(func() {
			envVars = map[string]string{
				"AWS_DEFAULT_REGION":    "us-west-2",
				"AWS_ACCESS_KEY_ID":     "some-access-key-id",
				"AWS_SECRET_ACCESS_KEY": "some-secret-access-key",
			}
		})

		const ErrTimeout = "10s"
		Context("no command line argument are provided", func() {
			It("should print a useful error", func() {
				session := start(nil, []string{}...)
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("specify one command of: down, show or up"))
			})
		})

		Context("when the action is unknown", func() {
			It("should print a useful error", func() {
				session := start(envVars, "-n", stackName, "nonsense_action")
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("Unknown command"))
				Expect(session.Err.Contents()).To(ContainSubstring("specify one command of: down, show or up"))
			})
		})

		Context("when required env vars are missing", func() {
			It("should print a useful error", func() {
				delete(envVars, "AWS_SECRET_ACCESS_KEY")

				session := start(envVars, "-n", stackName, "up")

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("missing .* AWS config"))
			})
		})

		Context("when the stack name is invalid", func() {
			It("should return a useful error", func() {
				session := start(envVars, "-n", "invalid_stack_name", "up")
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("invalid name: must match pattern"))
			})
		})

		Context("when application errors", func() {
			It("should inform the user", func() {
				envVars["AWS_SECRET_ACCESS_KEY"] = "some-invalid-key"

				session := start(envVars, "-n", stackName, "up")

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("AWS was not able to validate the provided access credentials"))
			})
		})
	})
})
