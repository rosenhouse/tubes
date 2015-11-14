package awsclient_test

import (
	"errors"
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Idempotent delete of a CloudFormation stack", func() {
	var (
		client               awsclient.Client
		cloudFormationClient *mocks.CloudFormationClient
		stackName            string
	)

	BeforeEach(func() {
		cloudFormationClient = &mocks.CloudFormationClient{}
		client = awsclient.Client{
			CloudFormation: cloudFormationClient,
		}
		stackName = fmt.Sprintf("some-stack-%x", rand.Int31()>>16)
	})

	It("should call DeleteStack", func() {
		Expect(client.DeleteStack(stackName)).To(Succeed())

		Expect(*cloudFormationClient.DeleteStackCall.Receives.Input.StackName).To(Equal(stackName))
	})

	Context("when AWS client returns an unrecognized error", func() {
		It("should return the error", func() {
			cloudFormationClient.DeleteStackCall.Returns.Error = errors.New("some error")

			Expect(client.DeleteStack(stackName)).To(MatchError("some error"))
		})
	})

})
