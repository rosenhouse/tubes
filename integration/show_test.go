package integration_test

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/rosenhouse/tubes/integration"
)

var _ = Describe("Show action", func() {
	var (
		stackName  string
		envVars    map[string]string
		workingDir string
		fakeAWS    *integration.FakeAWS
		start      func(args ...string) *gexec.Session

		manifestServer *httptest.Server
		boshIOServer   *httptest.Server
	)

	const NormalTimeout = "5s"

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

		session := start("-n", stackName, "up")
		Eventually(session, NormalTimeout).Should(gexec.Exit(0))
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

	It("should expose the SSH key", func() {
		session := start("-n", stackName, "show", "--ssh")

		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		pemBlock, _ := pem.Decode(session.Out.Contents())
		Expect(pemBlock).NotTo(BeNil())
		Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

		_, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should expose the director IP", func() {
		session := start("-n", stackName, "show", "--bosh-ip")

		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		Expect(session.Out.Contents()).To(Equal([]byte("192.168.12.13")))
	})

	It("should expose the bosh password", func() {
		session := start("-n", stackName, "show", "--bosh-password")

		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		Expect(session.Out.Contents()).To(HaveLen(12))
	})

	It("should expose the bosh settings as a sourcable environment file", func() {
		session := start("-n", stackName, "show", "--bosh-environment")

		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		envFile := string(session.Out.Contents())
		env := map[string]string{}

		re := regexp.MustCompile(`^export (\S*)="(\S*)"\s*$`)
		for _, line := range strings.Split(envFile, "\n") {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				env[matches[1]] = matches[2]
			}
		}
		Expect(env).To(HaveKeyWithValue("BOSH_TARGET", "192.168.12.13"))
		Expect(env).To(HaveKeyWithValue("BOSH_USER", "admin"))
		Expect(env).To(HaveKey("BOSH_PASSWORD"))
	})

	It("should support an explicit state directory, rather than the implicit subdirectory of the working directory", func() {
		defaultStateDir := filepath.Join(workingDir, "environments", stackName)
		session := start("-n", stackName, "--state-dir", defaultStateDir, "show", "--ssh")

		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		pemBlock, _ := pem.Decode(session.Out.Contents())
		Expect(pemBlock).NotTo(BeNil())
		Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))
	})

})
