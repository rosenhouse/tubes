package application_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Up", func() {
	BeforeEach(func() {
		awsClient.GetLatestNATBoxAMIIDCall.Returns.AMIID = "some-nat-box-ami-id"
		awsClient.GetBaseStackResourcesCall.Returns.Resources =
			awsclient.BaseStackResources{
				AccountID:        "ping pong",
				BOSHUser:         "some-bosh-user",
				NATInstanceID:    "some-nat-box-instance-id",
				NATElasticIP:     "some-nat-box-elastic-ip",
				VPCID:            "some-vpc-id",
				BOSHSubnetID:     "some-bosh-subnet-id",
				BOSHElasticIP:    "some-elastic-ip",
				AvailabilityZone: "some-availability-zone",
			}
		awsClient.CreateAccessKeyCall.Returns.AccessKey = "some-access-key"
		awsClient.CreateAccessKeyCall.Returns.SecretKey = "some-secret-key"

		awsClient.GetStackResourcesCalls = make([]mocks.GetStackResourcesCall, 1)
		awsClient.GetStackResourcesCalls[0].Returns.Resources = map[string]string{
			"ConcourseSecurityGroup": "some-concourse-security-group-id",
			"ConcourseSubnet":        "some-concourse-subnet-id",
			"LoadBalancer":           "some-concourse-elb",
		}
		manifestBuilder.BuildCall.Returns.AdminPassword = "some-bosh-password"
		cloudConfigGenerator.GenerateCall.Returns.Bytes = []byte("some-cloud-config")
	})

	It("should create a new ssh keypair", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(Equal(stackName))
	})

	It("should store the ssh keypair in the config store", func() {
		awsClient.CreateKeyPairCall.Returns.KeyPair = "some pem bytes"
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"ssh-key",
			[]byte("some pem bytes")))
	})

	It("should boot the base stack using the latest NAT ID", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(logBuffer).To(gbytes.Say("Creating keypair"))
		Expect(logBuffer).To(gbytes.Say("Looking for latest AWS NAT box AMI..."))
		Expect(logBuffer).To(gbytes.Say("Latest NAT box AMI is \"some-nat-box-ami-id\""))
		Expect(logBuffer).To(gbytes.Say("Upserting base stack.  Check CloudFormation console for details."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Generating BOSH init manifest"))
		Expect(logBuffer).To(gbytes.Say("Generating the concourse cloud config"))
		Expect(logBuffer).To(gbytes.Say("Finished"))

		Expect(awsClient.UpsertStackCalls[0].Receives.StackName).To(Equal(stackName + "-base"))
		Expect(awsClient.UpsertStackCalls[0].Receives.Template).To(Equal(awsclient.BaseStackTemplate.String()))
		Expect(awsClient.UpsertStackCalls[0].Receives.Parameters).To(Equal(map[string]string{
			"NATInstanceAMI": "some-nat-box-ami-id",
			"KeyName":        stackName,
		}))
	})

	It("should wait for the base stack to boot", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCalls[0].Receives.StackName).To(Equal(stackName + "-base"))
		Expect(awsClient.WaitForStackCalls[0].Receives.Pundit).To(Equal(awsclient.CloudFormationUpsertPundit{}))
	})

	It("should get the base stack resources", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(awsClient.GetBaseStackResourcesCall.Receives.StackName).To(Equal(stackName + "-base"))
	})

	It("should store the BOSH IP and NAT box IP in the config store", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"bosh-ip",
			[]byte("some-elastic-ip")))

		Expect(configStore.Values).To(HaveKeyWithValue(
			"nat-ip",
			[]byte("some-nat-box-elastic-ip")))
	})

	It("should create an access key for the BOSH user", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(awsClient.CreateAccessKeyCall.Receives.UserName).To(Equal("some-bosh-user"))
	})

	It("should provide the stack resources to the BOSH deployment manifest builder", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(manifestBuilder.BuildCall.Receives.StackName).To(Equal(stackName))
		Expect(manifestBuilder.BuildCall.Receives.Resources.AccountID).To(Equal("ping pong"))
		Expect(manifestBuilder.BuildCall.Receives.Resources.BOSHUser).To(Equal("some-bosh-user"))
		Expect(manifestBuilder.BuildCall.Receives.AccessKey).To(Equal("some-access-key"))
		Expect(manifestBuilder.BuildCall.Receives.SecretKey).To(Equal("some-secret-key"))
	})

	It("should store the BOSH deployment manifest", func() {
		manifestBuilder.BuildCall.Returns.ManifestYAML = []byte("some-manifest-bytes")

		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"director.yml",
			[]byte("some-manifest-bytes"),
		))
	})

	It("should store the BOSH password", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"bosh-password",
			[]byte("some-bosh-password"),
		))
	})

	It("should store a BOSH environment file", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"bosh-environment",
			[]byte(`export BOSH_TARGET="some-elastic-ip"
export BOSH_USER="admin"
export BOSH_PASSWORD="some-bosh-password"
export NAT_IP="some-nat-box-elastic-ip"`)))
	})

	It("should upsert the Concourse cloudformation stack", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(logBuffer).To(gbytes.Say("Upserting base stack.  Check CloudFormation console for details."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Upserting Concourse stack.  Check CloudFormation console for details."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Retrieving resource ids"))
		Expect(logBuffer).To(gbytes.Say("Generating the concourse cloud config"))
		Expect(logBuffer).To(gbytes.Say("Finished"))

		Expect(awsClient.UpsertStackCallCount).To(Equal(2))
		Expect(awsClient.UpsertStackCalls[1].Receives.StackName).To(Equal(stackName + "-concourse"))
		Expect(awsClient.UpsertStackCalls[1].Receives.Template).To(Equal(awsclient.ConcourseStackTemplate.String()))
		Expect(awsClient.UpsertStackCalls[1].Receives.Parameters).To(Equal(map[string]string{
			"VPCID":                    "some-vpc-id",
			"NATInstance":              "some-nat-box-instance-id",
			"PubliclyRoutableSubnetID": "some-bosh-subnet-id",
			"AvailabilityZone":         "some-availability-zone",
		}))
	})

	It("should wait for the Concourse stack to boot", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCallCount).To(Equal(2))
		Expect(awsClient.WaitForStackCalls[1].Receives.StackName).To(Equal(stackName + "-concourse"))
		Expect(awsClient.WaitForStackCalls[1].Receives.Pundit).To(Equal(awsclient.CloudFormationUpsertPundit{}))
	})

	It("should get the Concourse stack resources", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(awsClient.GetStackResourcesCalls[0].Receives.StackName).To(Equal(stackName + "-concourse"))
	})

	It("should generate the cloud config for concourse and store it", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(cloudConfigGenerator.GenerateCall.Receives.Resources).To(Equal(awsClient.GetStackResourcesCalls[0].Returns.Resources))
		Expect(configStore.Values["cloud-config.yml"]).To(Equal([]byte("some-cloud-config")))
	})

	Context("when the stackName contains invalid characters", func() {
		It("should immediately error", func() {
			Expect(app.Boot("invalid_name")).To(MatchError(fmt.Sprintf("invalid name: must match pattern %s", application.StackNamePattern)))
			Expect(logBuffer.Contents()).To(BeEmpty())
		})
	})

	Context("when the configStore is non-empty", func() {
		It("should immediately error", func() {
			configStore.Values["anything"] = []byte("hello")

			Expect(app.Boot(stackName)).To(MatchError("state directory must be empty"))
			Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(BeEmpty())
		})
	})

	Context("when an error arises from checking the config store for emptiness", func() {
		It("should immediately error", func() {
			configStore.IsEmptyError = errors.New("whatever")

			Expect(app.Boot(stackName)).To(MatchError("whatever"))
			Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(BeEmpty())
		})
	})

	Context("when getting the latest NAT AMI errors", func() {
		It("should immediately return the error", func() {
			awsClient.GetLatestNATBoxAMIIDCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.UpsertStackCalls).To(HaveLen(0))
		})
	})

	Context("when creating a keypair fails", func() {
		It("should immediately return the error", func() {
			awsClient.CreateKeyPairCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.UpsertStackCalls).To(HaveLen(0))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Looking for latest AWS NAT box AMI"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when storing the ssh key fails", func() {
		It("should return an error", func() {
			configStore.Errors["ssh-key"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Upserting base stack"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when upserting the base stack returns an error", func() {
		It("should immediately return the error", func() {
			awsClient.UpsertStackCalls = make([]mocks.UpsertStackCall, 1)
			awsClient.UpsertStackCalls[0].Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.WaitForStackCalls).To(BeEmpty())
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Stack update complete"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when waiting for the base stack returns an error", func() {
		It("should return the error", func() {
			awsClient.WaitForStackCalls = make([]mocks.WaitForStackCall, 1)
			awsClient.WaitForStackCalls[0].Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))

			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Stack update complete"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when getting the base stack resources fails", func() {
		It("should return the error", func() {
			awsClient.GetBaseStackResourcesCall.Returns.Error = errors.New("boom")

			Expect(app.Boot(stackName)).To(MatchError("boom"))
		})
	})

	Context("when storing the BOSH IP fails", func() {
		It("should return an error", func() {
			configStore.Errors["bosh-ip"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Generating BOSH init manifest"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when creating an access key fails", func() {
		It("should return the error", func() {
			awsClient.CreateAccessKeyCall.Returns.Error = errors.New("boom")

			Expect(app.Boot(stackName)).To(MatchError("boom"))
		})
	})

	Context("when building the BOSH director manifest yaml errors", func() {
		It("should return the error", func() {
			manifestBuilder.BuildCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
		})
	})

	Context("when storing the BOSH director manifest yaml fails", func() {
		It("should return an error", func() {
			configStore.Errors["director.yml"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Downloading the concourse manifest"))
		})
	})

	Context("when storing the BOSH password fails", func() {
		It("should return an error", func() {
			configStore.Errors["bosh-password"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Downloading the concourse manifest"))
		})
	})

	Context("when storing the BOSH environment fails", func() {
		It("should return an error", func() {
			configStore.Errors["bosh-environment"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Downloading the concourse manifest"))
		})
	})

	Context("when upserting the Concourse stack returns an error", func() {
		It("should immediately return the error", func() {
			awsClient.UpsertStackCalls = make([]mocks.UpsertStackCall, 2)
			awsClient.UpsertStackCalls[1].Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when waiting for the Concourse stack returns an error", func() {
		It("should return the error", func() {
			awsClient.WaitForStackCalls = make([]mocks.WaitForStackCall, 2)
			awsClient.WaitForStackCalls[1].Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))

			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when generating the cloud config fails", func() {
		It("should return an error", func() {
			cloudConfigGenerator.GenerateCall.Returns.Error = errors.New("potato")

			Expect(app.Boot(stackName)).To(MatchError("potato"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("potato"))
		})
	})

	Context("when storing the Concourse cloud config fails", func() {
		It("should return an error", func() {
			configStore.Errors["cloud-config.yml"] = errors.New("some-error")

			Expect(app.Boot(stackName)).To(MatchError("some-error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})
})
