package integration_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/rosenhouse/awsfaker"
)

var _ = Describe("Integration (mocking out AWS)", func() {
	var (
		stackName      string
		envVars        map[string]string
		workingDir     string
		fakeAWSBackend *FakeAWSBackend
		fakeAWS        *httptest.Server
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

		fakeAWSBackend = NewFakeAWSBackend(GinkgoWriter)
		fakeAWS = httptest.NewServer(awsfaker.New(awsfaker.Backend{EC2: fakeAWSBackend}))
		envVars = map[string]string{
			"AWS_DEFAULT_REGION":    "us-west-2",
			"AWS_ACCESS_KEY_ID":     "some-access-key-id",
			"AWS_SECRET_ACCESS_KEY": "some-secret-access-key",
			"TUBES_AWS_ENDPOINT":    fakeAWS.URL,
		}
	})

	AfterEach(func() {
		if fakeAWS != nil {
			fakeAWS.Close()
		}
	})

	Context("invalid user input", func() { // fast failing cases
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
				fakeAWSBackend.CreateKeyPairCall.ReturnsError = &awsfaker.ErrorResponse{
					AWSErrorCode:    "BadCredentials",
					AWSErrorMessage: "Your credentials are bad and you should feel bad",
					HTTPStatusCode:  http.StatusBadRequest,
				}

				session := start(envVars, "-n", stackName, "up")

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("credentials are bad"))
			})
		})
	})
})
