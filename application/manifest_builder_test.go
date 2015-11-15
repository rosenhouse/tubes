package application_test

import (
	"errors"
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/lib/director"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("ManifestBuilder", func() {
	var (
		directorManifestGenerator *mocks.DirectorManifestGenerator
		boshioClient              *mocks.BoshIOClient
		credentialsGenerator      *mocks.CredentialsGenerator
		baseStackResources        awsclient.BaseStackResources
		stackName                 string

		manifestBuilder *application.ManifestBuilder
	)

	BeforeEach(func() {
		directorManifestGenerator = &mocks.DirectorManifestGenerator{}
		boshioClient = mocks.NewBoshIOClient(2)
		credentialsGenerator = &mocks.CredentialsGenerator{}

		baseStackResources = awsclient.BaseStackResources{
			AvailabilityZone:  "some-availability-zone",
			BOSHSubnetCIDR:    "10.2.1.0/24",
			BOSHSubnetID:      "some-subnet-id",
			BOSHElasticIP:     "some-elastic-ip",
			BOSHSecurityGroup: "some-security-group",
		}
		stackName = fmt.Sprintf("some-stack-name-%x", rand.Int31())

		manifestBuilder = &application.ManifestBuilder{
			DirectorManifestGenerator: directorManifestGenerator,
			BoshIOClient:              boshioClient,
			CredentialsGenerator:      credentialsGenerator,
			AWSCredentials: director.AWSCredentials{
				Region:          "some-region",
				AccessKeyID:     "some-access-key",
				SecretAccessKey: "some-secret-key",
			},
		}

		boshioClient.LatestStemcellCall.Returns.Artifact.URL = "some-stemcell-url"
		boshioClient.LatestStemcellCall.Returns.Artifact.SHA = "some-stemcell-sha"
		boshioClient.LatestReleaseCalls[0].Returns.Artifact.URL = "some-aws-cpi-release-url"
		boshioClient.LatestReleaseCalls[0].Returns.Artifact.SHA = "some-aws-cpi-release-sha"
		boshioClient.LatestReleaseCalls[1].Returns.Artifact.URL = "some-bosh-director-release-url"
		boshioClient.LatestReleaseCalls[1].Returns.Artifact.SHA = "some-bosh-director-release-sha"

		credentialsGenerator.FillCallback = func(toFill interface{}) error {
			f := toFill.(*director.Credentials)
			f.MBus = "some-MBus-password"
			return nil
		}
	})

	Describe("configuring the software artifacts", func() {
		It("should discover the latest software", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshioClient.LatestStemcellCall.Receives.StemcellName).To(Equal("bosh-aws-xen-hvm-ubuntu-trusty-go_agent"))
			Expect(boshioClient.LatestReleaseCalls[0].Receives.ReleasePath).To(Equal("github.com/cloudfoundry-incubator/bosh-aws-cpi-release"))
			Expect(boshioClient.LatestReleaseCalls[1].Receives.ReleasePath).To(Equal("github.com/cloudfoundry/bosh"))
		})

		It("should pass the resulting software config to the director manifest generator", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			software := directorManifestGenerator.GenerateCall.Receives.Config.Software
			Expect(software.Stemcell).To(Equal(director.Artifact{
				URL: "some-stemcell-url",
				SHA: "some-stemcell-sha",
			}))
			Expect(software.BoshAWSCPIRelease).To(Equal(director.Artifact{
				URL: "some-aws-cpi-release-url",
				SHA: "some-aws-cpi-release-sha",
			}))
			Expect(software.BoshDirectorRelease).To(Equal(director.Artifact{
				URL: "some-bosh-director-release-url",
				SHA: "some-bosh-director-release-sha",
			}))
		})

		Context("when the boshio client errors", func() {
			It("should return stemcell errors", func() {
				boshioClient.LatestStemcellCall.Returns.Error = errors.New("some error")
				_, err := manifestBuilder.Build(stackName, baseStackResources)
				Expect(err).To(MatchError("some error"))
			})
			It("should return aws cpi release errors", func() {
				boshioClient.LatestReleaseCalls[0].Returns.Error = errors.New("some error")
				_, err := manifestBuilder.Build(stackName, baseStackResources)
				Expect(err).To(MatchError("some error"))
			})
			It("should return bosh director release errors", func() {
				boshioClient.LatestReleaseCalls[1].Returns.Error = errors.New("some error")
				_, err := manifestBuilder.Build(stackName, baseStackResources)
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("configuring bosh director credentials", func() {
		It("should generate new credentials", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)

			Expect(err).NotTo(HaveOccurred())
			credentials := directorManifestGenerator.GenerateCall.Receives.Config.Credentials

			Expect(credentials.MBus).To(Equal("some-MBus-password"))
		})
		Context("when the credential generation fails", func() {
			It("should return the error", func() {
				credentialsGenerator.FillCallback = func(toFill interface{}) error {
					return errors.New("filler error (ha ha)")
				}
				_, err := manifestBuilder.Build(stackName, baseStackResources)
				Expect(err).To(MatchError("filler error (ha ha)"))
			})
		})
	})

	Describe("configuring IPs and IDs", func() {
		It("should set the internal IP of the director to the CIDR base address + 6", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			internalIP := directorManifestGenerator.GenerateCall.Receives.Config.InternalIP
			Expect(internalIP).To(Equal("10.2.1.6"))
		})
		It("should work even with weird subnet sizes", func() {
			baseStackResources.BOSHSubnetCIDR = "10.0.0.128/25"
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			internalIP := directorManifestGenerator.GenerateCall.Receives.Config.InternalIP
			Expect(internalIP).To(Equal("10.0.0.134"))
		})
		It("should set the network config for AWS", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			awsConfig := directorManifestGenerator.GenerateCall.Receives.Config.AWSNetwork
			Expect(awsConfig.AvailabilityZone).To(Equal("some-availability-zone"))
			Expect(awsConfig.BOSHSubnetID).To(Equal("some-subnet-id"))
			Expect(awsConfig.BOSHSubnetCIDR).To(Equal("10.2.1.0/24"))
			Expect(awsConfig.ElasticIP).To(Equal("some-elastic-ip"))
			Expect(awsConfig.SecurityGroup).To(Equal("some-security-group"))
		})
	})

	Describe("configuring aws credentials", func() {
		It("should assume the ssh key name and path based on the stack name", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			awsSSHKey := directorManifestGenerator.GenerateCall.Receives.Config.AWSSSHKey
			Expect(awsSSHKey.Name).To(Equal(stackName))
			Expect(awsSSHKey.Path).To(Equal("./ssh-key"))
		})
		It("should set the region, access key and secret key", func() {
			_, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())

			awsCredentials := directorManifestGenerator.GenerateCall.Receives.Config.AWSCredentials
			Expect(awsCredentials).To(Equal(manifestBuilder.AWSCredentials))
		})
	})

	Describe("assembling the config into YAML", func() {
		It("should return the generated manifest as YAML bytes", func() {
			directorManifestGenerator.GenerateCall.Returns.Manifest.Name = "some-deployment-name"
			yamlBytes, err := manifestBuilder.Build(stackName, baseStackResources)
			Expect(err).NotTo(HaveOccurred())
			Expect(yamlBytes).To(ContainSubstring("name: some-deployment-name"))
		})

		Context("when generating the manifest errors", func() {
			It("should return the error", func() {
				directorManifestGenerator.GenerateCall.Returns.Error = errors.New("missing subnet")
				_, err := manifestBuilder.Build(stackName, baseStackResources)
				Expect(err).To(MatchError("missing subnet"))
			})
		})
	})
})
