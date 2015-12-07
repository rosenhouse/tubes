package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Destroy", func() {
	It("should get the stack resources to discover the BOSH user", func() {
		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(awsClient.GetBaseStackResourcesCall.Receives.StackName).To(Equal(stackName))
	})

	It("should delete the user's access keys", func() {
		awsClient.GetBaseStackResourcesCall.Returns.Resources.BOSHUser = "some-iam-user"
		awsClient.ListAccessKeysCall.Returns.AccessKeys = []string{"some-access-key"}

		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(awsClient.ListAccessKeysCall.Receives.UserName).To(Equal("some-iam-user"))
		Expect(awsClient.DeleteAccessKeyCall.Receives.UserName).To(Equal("some-iam-user"))
		Expect(awsClient.DeleteAccessKeyCall.Receives.AccessKey).To(Equal("some-access-key"))
	})

	It("should delete the Concourse stack and the base stack", func() {
		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(logBuffer).To(gbytes.Say("Inspecting stack"))
		Expect(logBuffer).To(gbytes.Say("Inspecting user"))
		Expect(logBuffer).To(gbytes.Say("Deleting access keys"))
		Expect(logBuffer).To(gbytes.Say("Deleting Concourse stack"))
		Expect(awsClient.DeleteStackCalls[0].Receives.StackName).To(Equal(stackName + "-concourse"))
		Expect(logBuffer).To(gbytes.Say("Delete complete"))
		Expect(logBuffer).To(gbytes.Say("Deleting base stack"))
		Expect(awsClient.DeleteStackCalls[1].Receives.StackName).To(Equal(stackName))
		Expect(logBuffer).To(gbytes.Say("Delete complete"))
		Expect(logBuffer).To(gbytes.Say("Deleting keypair"))
		Expect(logBuffer).To(gbytes.Say("Finished"))
	})
	It("should wait for the Concourse stack to be fully deleted", func() {
		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCalls[0].Receives.StackName).To(Equal(stackName + "-concourse"))
		Expect(awsClient.WaitForStackCalls[0].Receives.Pundit).To(Equal(awsclient.CloudFormationDeletePundit{}))
	})

	It("should wait for the base stack to be fully deleted", func() {
		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCalls[1].Receives.StackName).To(Equal(stackName))
		Expect(awsClient.WaitForStackCalls[1].Receives.Pundit).To(Equal(awsclient.CloudFormationDeletePundit{}))
	})

	It("should delete the ssh keypair", func() {
		Expect(app.Destroy(stackName)).To(Succeed())

		Expect(awsClient.DeleteKeyPairCall.Receives.StackName).To(Equal(stackName))
	})

	Context("when inspecting the stack fails", func() {
		It("should immediately return the error", func() {
			awsClient.GetBaseStackResourcesCall.Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
		})
	})

	Context("when getting the user's access keys fails", func() {
		It("should immediately return the error", func() {
			awsClient.ListAccessKeysCall.Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
		})
	})

	Context("when deleting the user's access keys fails", func() {
		It("should immediately return the error", func() {
			awsClient.ListAccessKeysCall.Returns.AccessKeys = []string{"some-key"}
			awsClient.DeleteAccessKeyCall.Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
		})
	})

	Context("when deleting the Concourse stack errors", func() {
		It("should immediately return the error", func() {
			awsClient.DeleteStackCalls = make([]mocks.DeleteStackCall, 1)
			awsClient.DeleteStackCalls[0].Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
			Expect(awsClient.WaitForStackCalls).To(BeEmpty())
		})
	})

	Context("when deleting the base stack errors", func() {
		It("should immediately return the error", func() {
			awsClient.DeleteStackCalls = make([]mocks.DeleteStackCall, 2)
			awsClient.DeleteStackCalls[1].Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
			Expect(awsClient.WaitForStackCalls).To(HaveLen(1))
		})
	})

	Context("when waiting for the base stack errors", func() {
		It("should return the error", func() {
			awsClient.WaitForStackCalls = make([]mocks.WaitForStackCall, 1)
			awsClient.WaitForStackCalls[0].Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
		})
	})

	Context("when deleting a keypair fails", func() {
		It("should immediately return the error", func() {
			awsClient.DeleteKeyPairCall.Returns.Error = errors.New("some error")

			Expect(app.Destroy(stackName)).To(MatchError("some error"))
		})
	})
})
