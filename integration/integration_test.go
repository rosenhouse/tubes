package integration_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/rosenhouse/tubes/integration"
)

var _ = Describe("Basic workflow", func() {
	var (
		stackName  string
		envVars    map[string]string
		workingDir string
		fakeAWS    *integration.FakeAWS
		start      func(args ...string) *gexec.Session

		manifestServer *httptest.Server
		boshIOServer   *httptest.Server
	)

	BeforeEach(func() {
		stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
		var err error
		workingDir, err = ioutil.TempDir("", "tubes-acceptance-test")
		Expect(err).NotTo(HaveOccurred())

		logger := integration.NewAWSCallLogger(GinkgoWriter)
		fakeAWS = integration.NewFakeAWS(logger)

		concourseManifestTemplate, err := ioutil.ReadFile("fixtures/concourse-template.yml")
		Expect(err).NotTo(HaveOccurred())
		manifestServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(concourseManifestTemplate)
		}))

		boshIOServer = httptest.NewServer(&integration.FakeBoshIO{})

		envVars = map[string]string{
			"AWS_DEFAULT_REGION":                    "us-west-2",
			"AWS_ACCESS_KEY_ID":                     "some-access-key-id",
			"AWS_SECRET_ACCESS_KEY":                 "some-secret-access-key",
			"TUBES_AWS_ENDPOINTS":                   fakeAWS.EndpointOverridesEnvVar(),
			"TUBES_CONCOURSE_MANIFEST_TEMPLATE_URL": manifestServer.URL + "/concourse-template.yml",
			"TUBES_BOSH_IO_URL":                     boshIOServer.URL,
		}

		start = buildStarter(&workingDir, envVars)
	})

	AfterEach(func() {
		fakeAWS.Close()

		if manifestServer != nil {
			manifestServer.Close()
		}

		if boshIOServer != nil {
			boshIOServer.Close()
		}
	})

	const NormalTimeout = "5s"

	It("should support basic environment manipulation", func() {

		By("booting a fresh environment", func() {
			session := start("-n", stackName, "up")

			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Creating keypair"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Looking for latest AWS NAT box AMI"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("ami-[a-f0-9]*"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting base stack"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Stack update complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Generating BOSH init manifest"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting Concourse stack"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
			Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		})

		By("exposing the director IP", func() {
			session := start("-n", stackName, "show", "--bosh-ip")

			Eventually(session, NormalTimeout).Should(gexec.Exit(0))

			Expect(session.Out.Contents()).To(Equal([]byte("192.168.12.13")))
		})

		By("exposing the bosh password", func() {
			session := start("-n", stackName, "show", "--bosh-password")

			Eventually(session, NormalTimeout).Should(gexec.Exit(0))

			Expect(session.Out.Contents()).To(HaveLen(12))
		})

		By("tearing down the environment", func() {
			session := start("-n", stackName, "down")

			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting Concourse stack"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Delete complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting base stack"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Delete complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting keypair"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
			Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		})
	})

	Context("invalid user input", func() { // fast failing cases
		const ErrTimeout = "10s"
		Context("no command line argument are provided", func() {
			It("should print a useful error", func() {
				session := start([]string{}...)
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("specify one command of: down, show or up"))
			})
		})

		Context("when the action is unknown", func() {
			It("should print a useful error", func() {
				session := start("-n", stackName, "nonsense_action")
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("Unknown command"))
				Expect(session.Err.Contents()).To(ContainSubstring("specify one command of: down, show or up"))
			})
		})
	})
})
