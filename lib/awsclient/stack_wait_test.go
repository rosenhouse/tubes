package awsclient_test

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("waiting for the stack changes to complete", func() {
	var (
		client                     awsclient.Client
		cloudFormationStatusPundit *mocks.CloudFormationStatusPundit
		cloudFormationClient       *mocks.CloudFormationClientMultiCall
		clock                      *mocks.Clock
		stackName                  string
		nCalls                     int
	)

	newResult := func(status string, err error) mocks.DescribeStacksCall {
		return mocks.DescribeStacksCall{
			Output: &cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					&cloudformation.Stack{
						StackStatus: aws.String(status),
					},
				},
			},
			Error: err,
		}
	}

	BeforeEach(func() {
		nCalls = rand.Intn(7) + 3

		cloudFormationClient = mocks.NewCloudFormationClientMultiCall(nCalls)
		cloudFormationStatusPundit = mocks.NewCloudFormationStatusPundit(nCalls)

		for i := 0; i < nCalls; i++ {
			cloudFormationClient.DescribeStacksCalls[i] = newResult(fmt.Sprintf("some status %d", i), nil)
			cloudFormationStatusPundit.IsHealthyCalls[i].Returns.Result = true
			cloudFormationStatusPundit.IsCompleteCalls[i].Returns.Result = false
		}
		cloudFormationStatusPundit.IsCompleteCalls[nCalls-1].Returns.Result = true
		clock = &mocks.Clock{}

		client = awsclient.Client{
			CloudFormation:             cloudFormationClient,
			CloudFormationStatusPundit: cloudFormationStatusPundit,
			Clock: clock,
			CloudFormationWaitTimeout: 10 * time.Minute,
		}
		stackName = fmt.Sprintf("some-stack-%x", rand.Int31()>>16)
	})

	It("should call DescribeStacks repeatedly", func() {
		Expect(client.WaitForStack(stackName)).To(Succeed())

		for i := 0; i < nCalls; i++ {
			Expect(*cloudFormationClient.DescribeStacksCalls[i].Input.StackName).To(Equal(stackName))
		}
	})

	It("should check each status with the pundit", func() {
		Expect(client.WaitForStack(stackName)).To(Succeed())

		for i := 0; i < nCalls; i++ {
			Expect(cloudFormationStatusPundit.IsHealthyCalls[i].Receives.StatusString).To(Equal(fmt.Sprintf("some status %d", i)))
			Expect(cloudFormationStatusPundit.IsCompleteCalls[i].Receives.StatusString).To(Equal(fmt.Sprintf("some status %d", i)))
		}
	})

	It("should sleep in between retries", func() {
		Expect(client.WaitForStack(stackName)).To(Succeed())

		for i := 0; i < nCalls-1; i++ {
			Expect(clock.SleepCalls[i].Receives.Duration).To(Equal(5 * time.Second))
		}
	})

	Context("when the pundit reports a status is not healthy", func() {
		BeforeEach(func() {
			cloudFormationClient.DescribeStacksCalls[1] = newResult("some bad status", nil)
			cloudFormationStatusPundit.IsHealthyCalls[1].Returns.Result = false
		})
		It("should abort and return an error", func() {
			Expect(client.WaitForStack(stackName)).To(MatchError(fmt.Sprintf("stack %q has unhealthy status %q", stackName, "some bad status")))
			Expect(cloudFormationStatusPundit.IsCompleteCalls[1].Receives.StatusString).To(BeEmpty())
			Expect(cloudFormationClient.DescribeStacksCalls[2].Input).To(BeNil())
		})
	})

	Context("when the pundit reports a status is healthy and complete", func() {
		BeforeEach(func() {
			cloudFormationClient.DescribeStacksCalls[1] = newResult("some complete status", nil)
			cloudFormationStatusPundit.IsCompleteCalls[1].Returns.Result = true
		})
		It("should return immediately", func() {
			Expect(client.WaitForStack(stackName)).To(Succeed())
			Expect(cloudFormationClient.DescribeStacksCalls[2].Input).To(BeNil())
			Expect(cloudFormationStatusPundit.IsHealthyCalls[2].Receives.StatusString).To(BeEmpty())
		})
	})

	Context("when the stack change doesn't complete within the timeout", func() {
		It("should return an error", func() {
			nCalls = 15
			cloudFormationClient = mocks.NewCloudFormationClientMultiCall(nCalls)
			cloudFormationStatusPundit = mocks.NewCloudFormationStatusPundit(nCalls)

			for i := 0; i < nCalls; i++ {
				cloudFormationClient.DescribeStacksCalls[i] = newResult(fmt.Sprintf("some status %d", i), nil)
				cloudFormationStatusPundit.IsHealthyCalls[i].Returns.Result = true
				cloudFormationStatusPundit.IsCompleteCalls[i].Returns.Result = false
			}
			client = awsclient.Client{
				CloudFormation:             cloudFormationClient,
				CloudFormationStatusPundit: cloudFormationStatusPundit,
				Clock: clock,
				CloudFormationWaitTimeout: 65 * time.Second,
			}

			Expect(client.WaitForStack(stackName)).To(MatchError(
				"timed out waiting for stack change to complete (max 1m5s, some status 13)"))
		})
	})

	Context("when the DescribeStacks call errors", func() {
		It("should immediately return the error", func() {
			cloudFormationClient.DescribeStacksCalls[1] = newResult("whatever", errors.New("some aws error"))

			Expect(client.WaitForStack(stackName)).To(MatchError("some aws error"))
			Expect(cloudFormationClient.DescribeStacksCalls[2].Input).To(BeNil())
		})
	})
})
