package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("Destroy", func() {
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
