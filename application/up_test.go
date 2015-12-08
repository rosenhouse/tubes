package application_test

import (
	"errors"
	"fmt"
	"math/rand"

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
				VPCID:            "some-vpc-id",
				BOSHSubnetID:     "some-bosh-subnet-id",
				BOSHElasticIP:    "some-elastic-ip",
				AvailabilityZone: "some-availability-zone",
			}
		awsClient.CreateAccessKeyCall.Returns.AccessKey = "some-access-key"
		awsClient.CreateAccessKeyCall.Returns.SecretKey = "some-secret-key"

		credentialsGenerator.FillCallback = func(toFill interface{}) error {
			f := toFill.(*application.ConcourseCredentials)
			f.DBPassword = "some-db-password"
			return nil
		}
		awsClient.GetStackResourcesCalls = make([]mocks.GetStackResourcesCall, 1)
		awsClient.GetStackResourcesCalls[0].Returns.Resources = map[string]string{
			"ConcourseSecurityGroup": "some-concourse-security-group-id",
			"ConcourseSubnet":        "some-concourse-subnet-id",
			"LoadBalancer":           "some-concourse-elb",
		}
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
		Expect(logBuffer).To(gbytes.Say("Downloading the concourse manifest from " + app.ConcourseTemplateURL))
		Expect(logBuffer).To(gbytes.Say("Generating the concourse manifest"))
		Expect(logBuffer).To(gbytes.Say("Finished"))

		Expect(awsClient.UpsertStackCalls[0].Receives.StackName).To(Equal(stackName))
		Expect(awsClient.UpsertStackCalls[0].Receives.Template).To(Equal(awsclient.BaseStackTemplate.String()))
		Expect(awsClient.UpsertStackCalls[0].Receives.Parameters).To(Equal(map[string]string{
			"NATInstanceAMI": "some-nat-box-ami-id",
			"KeyName":        stackName,
		}))
	})

	It("should wait for the base stack to boot", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCalls[0].Receives.StackName).To(Equal(stackName))
		Expect(awsClient.WaitForStackCalls[0].Receives.Pundit).To(Equal(awsclient.CloudFormationUpsertPundit{}))
	})

	It("should get the base stack resources", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(awsClient.GetBaseStackResourcesCall.Receives.StackName).To(Equal(stackName))
	})

	It("should store the BOSH IP in the config store", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"bosh-ip",
			[]byte("some-elastic-ip")))
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
		manifestBuilder.BuildCall.Returns.AdminPassword = "some-bosh-password"

		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"bosh-password",
			[]byte("some-bosh-password"),
		))
	})

	It("should upsert the Concourse cloudformation stack", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(logBuffer).To(gbytes.Say("Upserting base stack.  Check CloudFormation console for details."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Upserting Concourse stack.  Check CloudFormation console for details."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Retrieving resource ids"))
		Expect(logBuffer).To(gbytes.Say("Generating the concourse manifest"))
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

	It("should get the concourse manifest template", func() {
		manifestTemplate := fmt.Sprintf("some-manifest-bytes-%x", rand.Int31())
		httpClient.GetCall.Returns.Body = []byte(manifestTemplate + " REPLACE_WITH_AVAILABILITY_ZONE ")
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring(manifestTemplate))

		Expect(httpClient.GetCall.Receives.Path).To(Equal(app.ConcourseTemplateURL))
	})

	It("should replace all placeholders in the concourse manifest template", func() {
		httpClient.GetCall.Returns.Body = []byte(" REPLACE_WITH_AVAILABILITY_ZONE" +
			" REPLACE_WITH_UUID" +
			" REPLACE_WITH_DB_PASSWORD" +
			" REPLACE_WITH_INTERNAL_SECURITY_GROUP_NAME" +
			" REPLACE_WITH_INTERNAL_SUBNET" +
			" REPLACE_WITH_WEB_ELB_NAME")
		awsClient.GetBaseStackResourcesCall.Returns.Resources.AWSRegion = "some-region"
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(configStore.Values["concourse.yml"]).NotTo(ContainSubstring("REPLACE_WITH_"))
	})

	It("should replace placeholders n the concourse manifest template with the correct values", func() {
		httpClient.GetCall.Returns.Body = []byte(
			"availability_zone: &az REPLACE_WITH_AVAILABILITY_ZONE" +
				" password: REPLACE_WITH_DB_PASSWORD" +
				" director_uuid: REPLACE_WITH_UUID" +
				" security_groups: [REPLACE_WITH_INTERNAL_SECURITY_GROUP_NAME]" +
				" subnet: REPLACE_WITH_INTERNAL_SUBNET" +
				" elbs: [REPLACE_WITH_WEB_ELB_NAME]")
		awsClient.GetBaseStackResourcesCall.Returns.Resources.AWSRegion = "some-region"
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("availability_zone: &az some-availability-zone"))
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("password: some-db-password"))
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("security_groups: [some-concourse-security-group-id]"))
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("subnet: some-concourse-subnet"))
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("elbs: [some-concourse-elb]"))
		Expect(configStore.Values["concourse.yml"]).To(ContainSubstring("director_uuid: YOUR_DIRECTOR_UUID_HERE"))
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

	Context("when downloading the concourse manifest template returns an error", func() {
		It("should return an error", func() {
			httpClient.GetCall.Returns.Error = errors.New("Bad Request")

			Expect(app.Boot(stackName)).To(MatchError("Bad Request"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Generating the concourse manifest"))
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

	Context("when storing the Concourse manifest yaml fails", func() {
		It("should return an error", func() {
			configStore.Errors["concourse.yml"] = errors.New("some concourse manifest storage error")

			Expect(app.Boot(stackName)).To(MatchError("some concourse manifest storage error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when generating random credentials fails", func() {
		It("should return an error", func() {
			credentialsGenerator.FillCallback = func(toFill interface{}) error {
				return errors.New("filler error (ha ha)")
			}

			Expect(app.Boot(stackName)).To(MatchError("filler error (ha ha)"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})
})
