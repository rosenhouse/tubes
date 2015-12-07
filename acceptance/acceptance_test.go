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

		It("should support basic environment manipulation", func() { // slow happy path
			const NormalTimeout = "10s"
			const StackChangeTimeout = "6m"

			By("booting a fresh environment", func() {
				session := start(envVars, "-n", stackName, "up")

				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Creating keypair"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Looking for latest AWS NAT box AMI"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("ami-[a-f0-9]*"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting base stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Stack update complete"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting Concourse stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Stack update complete"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
				Eventually(session, NormalTimeout).Should(gexec.Exit(0))
			})

			defaultStateDir := filepath.Join(workingDir, "environments", stackName)
			By("storing the SSH key on the filesystem", func() {
				Expect(ioutil.ReadFile(filepath.Join(defaultStateDir, "ssh-key"))).To(ContainSubstring("RSA PRIVATE KEY"))
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

			By("supporting an explicit state directory, rather than the implicit subdirectory of the working directory", func() {
				session := start(envVars, "-n", stackName, "--state-dir", defaultStateDir, "show")

				Eventually(session, NormalTimeout).Should(gexec.Exit(0))

				pemBlock, _ := pem.Decode(session.Out.Contents())
				Expect(pemBlock).NotTo(BeNil())
				Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))
			})

			By("storing a generated BOSH director manifest in the state directory", func() {
				directorYAMLBytes, err := ioutil.ReadFile(filepath.Join(defaultStateDir, "director.yml"))
				Expect(err).NotTo(HaveOccurred())

				Expect(directorYAMLBytes).To(ContainSubstring("resource_pools:"))

				By("ensuring we create fresh credentials for the BOSH director")
				Expect(directorYAMLBytes).NotTo(ContainSubstring(envVars["AWS_SECRET_ACCESS_KEY"]))
			})

			By("tearing down the environment", func() {
				session := start(envVars, "-n", stackName, "down")

				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting Concourse stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Delete complete"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting base stack"))
				Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Delete complete"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting keypair"))
				Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
				Eventually(session, NormalTimeout).Should(gexec.Exit(0))
			})
		})
	})
})
