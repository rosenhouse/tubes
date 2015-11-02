package director_test

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/manifests"

	. "github.com/rosenhouse/tubes/lib/director"
	. "github.com/rosenhouse/tubes/lib/matchers"
)

var _ = Describe("Generating a deployment manifest for a BOSH Director", func() {

	var (
		expectedManifestString string
		expectedManifest       manifests.Manifest
		softwareConfig         Software
		awsConfig              AWSConfig
		director               Director
	)

	BeforeEach(func() {
		softwareConfig = Software{
			BoshDirectorRelease: Artifact{
				URL: "https://bosh.io/d/github.com/cloudfoundry/bosh?v=219",
				SHA: "bbd03790a2839aab26d3fa4cfe1493d361872f33",
			},
			BoshAWSCPIRelease: Artifact{
				URL: "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=35",
				SHA: "2d51f151f99d59e43fa50b585599d32bcc72e297",
			},
			Stemcell: Artifact{
				URL: "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3012",
				SHA: "3380b55948abe4c437dee97f67d2d8df4eec3fc1",
			},
		}
		awsConfig = AWSConfig{
			InstanceType:     "m3.xlarge",
			AvailabilityZone: "AVAILABILITY-ZONE",
			BOSHSubnet: AWSSubnet{
				CIDR:     "10.0.0.0/24",
				SubnetID: "SUBNET-ID",
			},
			ElasticIP:       "ELASTIC-IP",
			PrivateKeyPath:  "./bosh.pem",
			AccessKeyID:     "ACCESS-KEY-ID",
			SecretAccessKey: "SECRET-ACCESS-KEY",
			PrivateKeyName:  "bosh",
			SecurityGroup:   "bosh",
			Region:          "us-east-1",
		}
		director = Director{
			Software:  softwareConfig,
			AWSConfig: awsConfig,
			Credentials: Credentials{
				MBus:              "mbus-password",
				NATS:              "nats-password",
				Redis:             "redis-password",
				Postgres:          "postgres-password",
				Registry:          "admin",
				BlobstoreDirector: "director-password",
				BlobstoreAgent:    "agent-password",
				HM:                "hm-password",
				Admin:             "admin",
			},
			InternalIP: "10.0.0.6",
		}

		expectedManifestBytes, err := ioutil.ReadFile("fixtures/bosh-init-aws.yml")
		Expect(err).NotTo(HaveOccurred())
		expectedManifestString = string(expectedManifestBytes)
		expectedManifest = manifests.Manifest{}
		Expect(yaml.Unmarshal(expectedManifestBytes, &expectedManifest)).To(Succeed())
	})

	Describe("equality of structural data", func() {
		It("should set the fields correctly", func() {
			actualManifest, err := director.Generate()
			Expect(err).NotTo(HaveOccurred())

			Expect(actualManifest.Name).To(Equal(expectedManifest.Name))
			Expect(actualManifest.Releases).To(Equal(expectedManifest.Releases))
			Expect(actualManifest.ResourcePools).To(Equal(expectedManifest.ResourcePools))
			Expect(actualManifest.DiskPools).To(Equal(expectedManifest.DiskPools))
			Expect(actualManifest.Networks).To(Equal(expectedManifest.Networks))
			Expect(actualManifest.CloudProvider).To(Equal(expectedManifest.CloudProvider))
			Expect(actualManifest.Jobs).To(Equal(expectedManifest.Jobs))
		})

		It("should match the entire structure", func() {
			actualManifest, err := director.Generate()
			Expect(err).NotTo(HaveOccurred())

			Expect(actualManifest).To(Equal(expectedManifest))
		})
	})

	Describe("equality of serialized data", func() {
		It("should have all the same data as the fixture", func() {
			actualManifest, err := director.Generate()
			Expect(err).NotTo(HaveOccurred())
			actualString := actualManifest.String()

			Expect(actualString).To(MatchYAML(expectedManifestString))
		})
	})
})
