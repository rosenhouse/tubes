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

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/rosenhouse/tubes/integration"
)

var _ = Describe("Integration (mocking out AWS)", func() {
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

		defaultStateDir := filepath.Join(workingDir, "environments", stackName)
		By("storing the SSH key on the filesystem", func() {
			Expect(ioutil.ReadFile(filepath.Join(defaultStateDir, "ssh-key"))).To(ContainSubstring("RSA PRIVATE KEY"))
		})
		By("storing the BOSH IP on the filesystem", func() {
			Expect(ioutil.ReadFile(filepath.Join(defaultStateDir, "bosh-ip"))).To(Equal([]byte("192.168.12.13")))
		})

		By("exposing the SSH key", func() {
			session := start("-n", stackName, "show", "--ssh")

			Eventually(session, NormalTimeout).Should(gexec.Exit(0))

			pemBlock, _ := pem.Decode(session.Out.Contents())
			Expect(pemBlock).NotTo(BeNil())
			Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

			_, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
			Expect(err).NotTo(HaveOccurred())
		})

		By("exposing the director IP", func() {
			session := start("-n", stackName, "show", "--bosh-ip")

			Eventually(session, NormalTimeout).Should(gexec.Exit(0))

			Expect(session.Out.Contents()).To(Equal([]byte("192.168.12.13")))
		})

		By("supporting an explicit state directory, rather than the implicit subdirectory of the working directory", func() {
			session := start("-n", stackName, "--state-dir", defaultStateDir, "show", "--ssh")

			Eventually(session, NormalTimeout).Should(gexec.Exit(0))

			pemBlock, _ := pem.Decode(session.Out.Contents())
			Expect(pemBlock).NotTo(BeNil())
			Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))
		})

		By("storing a generated BOSH director manifest in the state directory", func() {
			directorYAMLBytes, err := ioutil.ReadFile(filepath.Join(defaultStateDir, "director.yml"))
			Expect(err).NotTo(HaveOccurred())

			var directorYAML struct {
				Networks []struct {
					Subnets []struct {
						Cloud_Properties struct {
							Subnet string
						}
					}
				}
				Cloud_Provider struct {
					MBus       string
					SSH_Tunnel struct {
						Host string
					}
					Properties struct {
						AWS struct {
							Access_Key_ID           string
							Secret_Access_Key       string
							Default_Security_Groups []string
						}
					}
				}
			}
			Expect(yaml.Unmarshal(directorYAMLBytes, &directorYAML)).To(Succeed())

			Expect(directorYAML.Networks[0].Subnets[0].Cloud_Properties.Subnet).To(Equal("subnet-12345"))
			Expect(directorYAML.Cloud_Provider.SSH_Tunnel.Host).To(Equal("192.168.12.13"))
			Expect(directorYAML.Cloud_Provider.MBus).To(ContainSubstring("@192.168.12.13:6868"))
			Expect(directorYAML.Cloud_Provider.Properties.AWS.Access_Key_ID).To(Equal("some-access-key"))
			Expect(directorYAML.Cloud_Provider.Properties.AWS.Secret_Access_Key).To(Equal("some-secret-key"))
			Expect(directorYAML.Cloud_Provider.Properties.AWS.Default_Security_Groups[0]).To(Equal("sg-1234"))

			By("ensuring we create fresh credentials for the BOSH director")
			Expect(directorYAMLBytes).NotTo(ContainSubstring(envVars["AWS_SECRET_ACCESS_KEY"]))
		})

		By("storing a generated Concourse deployment manifest in the state directory", func() {
			concourseYAMLBytes, err := ioutil.ReadFile(filepath.Join(defaultStateDir, "concourse.yml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(concourseYAMLBytes).To(ContainSubstring("network: concourse"))
			Expect(concourseYAMLBytes).To(ContainSubstring("us-west-2"))
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

	It("should create a CloudFormation stack for the BOSH director", func() {
		Expect(fakeAWS.CloudFormation.Stacks).To(HaveLen(0))
		session := start("-n", stackName, "up")

		Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting base stack"))
		Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Stack update complete"))
		Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		Expect(*fakeAWS.CloudFormation.Stacks[0].StackName).To(Equal(stackName))
	})

	It("should create a CloudFormation stack for Concourse", func() {
		session := start("-n", stackName, "up")
		Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting Concourse stack"))
		Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		Expect(fakeAWS.CloudFormation.Stacks).To(HaveLen(2))
		Expect(*fakeAWS.CloudFormation.Stacks[1].StackName).To(Equal(stackName + "-concourse"))
	})

	It("should generate a Concourse manifest without any template placeholders", func() {
		session := start("-n", stackName, "up")
		Eventually(session, NormalTimeout).Should(gexec.Exit(0))

		defaultStateDir := filepath.Join(workingDir, "environments", stackName)
		concourseYAMLBytes, err := ioutil.ReadFile(filepath.Join(defaultStateDir, "concourse.yml"))
		Expect(err).NotTo(HaveOccurred())

		Expect(concourseYAMLBytes).NotTo(ContainSubstring("REPLACE_WITH_"))
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

		Context("when required env vars are missing", func() {
			It("should print a useful error", func() {
				delete(envVars, "AWS_SECRET_ACCESS_KEY")

				session := start("-n", stackName, "up")

				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("missing .* AWS config"))
			})
		})

		Context("when the stack name is invalid", func() {
			It("should return a useful error", func() {
				session := start("-n", "invalid_stack_name", "up")
				Eventually(session, ErrTimeout).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("invalid name: must match pattern"))
			})
		})
	})
})
