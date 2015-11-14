package application_test

import (
	"errors"
	"fmt"
	"log"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Application", func() {

	var (
		awsClient *mocks.AWSClient

		app *application.Application

		stackName string
		logBuffer *gbytes.Buffer
	)

	BeforeEach(func() {
		awsClient = &mocks.AWSClient{}

		logBuffer = gbytes.NewBuffer()

		app = &application.Application{
			AWSClient: awsClient,
			Logger:    log.New(logBuffer, "", 0),
		}

		awsClient.GetLatestNATBoxAMIIDCall.Returns.AMIID = "some-nat-box-ami-id"

		stackName = fmt.Sprintf("some-stack-name-%x", rand.Int31())
	})

	Describe("Boot", func() {
		It("should boot the base stack using the latest NAT ID", func() {
			Expect(app.Boot(stackName)).To(Succeed())

			Expect(logBuffer).To(gbytes.Say("Looking for latest AWS NAT box AMI..."))
			Expect(logBuffer).To(gbytes.Say("Latest NAT box AMI is \"some-nat-box-ami-id\""))
			Expect(logBuffer).To(gbytes.Say("Creating keypair"))
			Expect(logBuffer).To(gbytes.Say("Upserting stack..."))
			Expect(logBuffer).To(gbytes.Say("Finished"))

			Expect(awsClient.UpsertStackCall.Receives.StackName).To(Equal(stackName))
			Expect(awsClient.UpsertStackCall.Receives.Template).To(Equal(awsclient.BaseStackTemplate.String()))
			Expect(awsClient.UpsertStackCall.Receives.Parameters).To(Equal(map[string]string{
				"NATInstanceAMI": "some-nat-box-ami-id",
				"KeyName":        stackName,
			}))
		})

		It("should wait for the stack to boot", func() {
			Expect(app.Boot(stackName)).To(Succeed())

			Expect(awsClient.WaitForStackCall.Receives.StackName).To(Equal(stackName))
			Expect(awsClient.WaitForStackCall.Receives.Pundit).To(Equal(awsclient.CloudFormationUpsertPundit{}))
		})

		It("should create a new ssh keypair", func() {
			Expect(app.Boot(stackName)).To(Succeed())

			Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(Equal(stackName))
		})

		Context("when the stackName contains invalid characters", func() {
			It("should immediately error", func() {
				Expect(app.Boot("invalid_name")).To(MatchError(fmt.Sprintf("invalid name: must match pattern %s", application.StackNamePattern)))
				Expect(logBuffer.Contents()).To(BeEmpty())
			})
		})

		Context("when getting the latest NAT AMI errors", func() {
			It("should immediately return the error", func() {
				awsClient.GetLatestNATBoxAMIIDCall.Returns.Error = errors.New("some error")

				Expect(app.Boot(stackName)).To(MatchError("some error"))
				Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(BeEmpty())
				Expect(awsClient.UpsertStackCall.Receives.StackName).To(BeEmpty())
				Expect(awsClient.WaitForStackCall.Receives.StackName).To(BeEmpty())
			})
		})

		Context("when creating a keypair fails", func() {
			It("should immediately return the error", func() {
				awsClient.CreateKeyPairCall.Returns.Error = errors.New("some error")

				Expect(app.Boot(stackName)).To(MatchError("some error"))
				Expect(awsClient.UpsertStackCall.Receives.StackName).To(BeEmpty())
			})
		})

		Context("when upserting the stack errors", func() {
			It("should immediately return the error", func() {
				awsClient.UpsertStackCall.Returns.Error = errors.New("some error")

				Expect(app.Boot(stackName)).To(MatchError("some error"))
				Expect(awsClient.WaitForStackCall.Receives.StackName).To(BeEmpty())
			})
		})

		Context("when waiting for the stack errors", func() {
			It("should return the error", func() {
				awsClient.WaitForStackCall.Returns.Error = errors.New("some error")

				Expect(app.Boot(stackName)).To(MatchError("some error"))
			})
		})
	})

	Describe("Destroy", func() {
		It("should delete the stack", func() {
			Expect(app.Destroy(stackName)).To(Succeed())

			Expect(awsClient.DeleteStackCall.Receives.StackName).To(Equal(stackName))

			Expect(logBuffer).To(gbytes.Say("Deleting stack"))
			Expect(logBuffer).To(gbytes.Say("Delete complete"))
			Expect(logBuffer).To(gbytes.Say("Deleting keypair"))
			Expect(logBuffer).To(gbytes.Say("Finished"))
		})

		It("should wait for the stack be fully deleted", func() {
			Expect(app.Destroy(stackName)).To(Succeed())

			Expect(awsClient.WaitForStackCall.Receives.StackName).To(Equal(stackName))
			Expect(awsClient.WaitForStackCall.Receives.Pundit).To(Equal(awsclient.CloudFormationDeletePundit{}))
		})

		It("should delete the ssk keypair", func() {
			Expect(app.Destroy(stackName)).To(Succeed())

			Expect(awsClient.DeleteKeyPairCall.Receives.StackName).To(Equal(stackName))
		})

		Context("when deleting a keypair fails", func() {
			It("should immediately return the error", func() {
				awsClient.DeleteKeyPairCall.Returns.Error = errors.New("some error")

				Expect(app.Destroy(stackName)).To(MatchError("some error"))
			})
		})

		Context("when deleting the stack errors", func() {
			It("should immediately return the error", func() {
				awsClient.DeleteStackCall.Returns.Error = errors.New("some error")

				Expect(app.Destroy(stackName)).To(MatchError("some error"))
				Expect(awsClient.WaitForStackCall.Receives.StackName).To(BeEmpty())
			})
		})

		Context("when waiting for the stack errors", func() {
			It("should return the error", func() {
				awsClient.WaitForStackCall.Returns.Error = errors.New("some error")

				Expect(app.Destroy(stackName)).To(MatchError("some error"))
			})
		})
	})
})
