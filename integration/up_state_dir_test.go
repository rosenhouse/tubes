package integration_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/rosenhouse/tubes/integration"
)

var _ = Describe("Up action dependency on state directory", func() {
	var (
		stackName  string
		envVars    map[string]string
		workingDir string
		fakeAWS    *integration.FakeAWS
		start      func(args ...string) *gexec.Session
		args       []string
	)

	BeforeEach(func() {
		stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
		var err error
		workingDir, err = ioutil.TempDir("", "tubes-acceptance-test")
		Expect(err).NotTo(HaveOccurred())

		logger := integration.NewAWSCallLogger(GinkgoWriter)
		fakeAWS = integration.NewFakeAWS(logger)

		envVars = map[string]string{
			"AWS_DEFAULT_REGION":    "us-west-2",
			"AWS_ACCESS_KEY_ID":     "some-access-key-id",
			"AWS_SECRET_ACCESS_KEY": "some-secret-access-key",
			"TUBES_AWS_ENDPOINTS":   fakeAWS.EndpointOverridesEnvVar(),
		}

		start = buildStarter(&workingDir, envVars)
	})

	AfterEach(func() {
		fakeAWS.Close()
	})

	var dirExists = func(path string) bool {
		s, e := os.Stat(path)
		if e == nil {
			if s.IsDir() {
				return true
			}
			panic("file exists, but not a directory")
		} else {
			if os.IsNotExist(e) {
				return false
			}
			panic("unexpected error: " + e.Error())
		}
	}
	const DefaultTimeout = "5s"

	Context("when the state directory is set explicitly", func() {

		var stateDir string
		BeforeEach(func() {
			var err error
			stateDir, err = ioutil.TempDir("", "tubes-integration-state-dir")
			Expect(err).NotTo(HaveOccurred())
			args = []string{"--state-dir", stateDir, "-n", stackName, "up"}
		})

		Context("when the state directory does not exist", func() {
			It("should error", func() {
				stateDir = stateDir + "-nope"
				args = []string{"--state-dir", stateDir, "-n", stackName, "up"}
				Expect(dirExists(stateDir)).To(BeFalse())
				session := start(args...)
				Eventually(session, DefaultTimeout).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("state directory not found"))
			})
		})

		Context("when the state directory does exist", func() {
			BeforeEach(func() {
				Expect(dirExists(stateDir)).To(BeTrue())
			})
			Context("when the state directory is empty", func() {
				BeforeEach(func() {
					Expect(ioutil.ReadDir(stateDir)).To(HaveLen(0))
				})
				It("should create a new stack", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(0))
					Expect(*fakeAWS.CloudFormation.Stacks[0].StackStatus).To(Equal("CREATE_COMPLETE"))
				})
				It("should save the new state to the state directory", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(0))
					Expect(ioutil.ReadDir(stateDir)).To(HaveLen(2))
				})
			})
			Context("when the state directory is not empty", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(stateDir, "anything"), nil, 0600)).To(Succeed())
				})
				It("should error", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say("state directory must be empty"))
				})
			})
		})
	})

	Context("when the state directory is not set", func() {

		var implicitStateDir string
		BeforeEach(func() {
			args = []string{"-n", stackName, "up"}
			implicitStateDir = filepath.Join(workingDir, "environments", stackName)
		})

		Context("when the working directory does not already contain the implied state dir", func() {
			BeforeEach(func() {
				Expect(dirExists(implicitStateDir)).To(BeFalse())
			})
			It("should create a new state dir at the implied path and populate it", func() {
				session := start(args...)
				Eventually(session, DefaultTimeout).Should(gexec.Exit(0))
				Expect(dirExists(implicitStateDir)).To(BeTrue())
				Expect(ioutil.ReadDir(implicitStateDir)).To(HaveLen(2))
			})
		})

		Context("when the working directory already contains a state directory at the implied path", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(implicitStateDir, 0777)).To(Succeed())
			})
			Context("when that implied state directory is empty", func() {
				BeforeEach(func() {
					Expect(ioutil.ReadDir(implicitStateDir)).To(HaveLen(0))
				})
				It("should create a new stack", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(0))
					Expect(*fakeAWS.CloudFormation.Stacks[0].StackStatus).To(Equal("CREATE_COMPLETE"))
				})
				It("should save the new state to the implied state directory", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(0))
					Expect(ioutil.ReadDir(implicitStateDir)).To(HaveLen(2))
				})
			})
			Context("when that implied state directory is not empty", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(implicitStateDir, "anything"), nil, 0600)).To(Succeed())
				})
				It("should error", func() {
					session := start(args...)
					Eventually(session, DefaultTimeout).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say("state directory must be empty"))
				})
			})
		})
	})
})
