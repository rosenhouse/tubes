package awsclient_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Idempotent upsert of a CloudFormation stack", func() {
	var (
		client               awsclient.Client
		cloudFormationClient *mocks.CloudFormationClient
		stackName            string
		template             string
		parameters           map[string]string
	)

	BeforeEach(func() {
		cloudFormationClient = &mocks.CloudFormationClient{}
		client = awsclient.Client{
			CloudFormation: cloudFormationClient,
		}
		stackName = fmt.Sprintf("some-stack-%x", rand.Int31()>>16)
		template = fmt.Sprintf(`{ "some": "template" }`)
		parameters = map[string]string{"a": "b", "c": "d", "e": "f"}

		cloudFormationClient.DescribeStacksCall.Returns.Output = &cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String("nonsense status"),
				},
			},
		}
	})

	It("should call DescribeStack to determine if the stack already exists", func() {
		client.UpsertStack(stackName, template, parameters)

		Expect(*cloudFormationClient.DescribeStacksCall.Receives.Input.StackName).To(Equal(stackName))
	})

	Context("when the stack does not yet exist", func() {
		BeforeEach(func() {
			cloudFormationClient.DescribeStacksCall.Returns.Output = nil
			cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(
				awserr.New("ValidationError", "Stack with id STACKNAMEHERE does not exist", nil),
				400, "some-request-id")
		})

		It("should try to create the stack, not update it", func() {
			Expect(client.UpsertStack(stackName, template, parameters)).To(Succeed())

			Expect(*cloudFormationClient.CreateStackCall.Receives.Input.StackName).To(Equal(stackName))
			Expect(*cloudFormationClient.CreateStackCall.Receives.Input.TemplateBody).To(Equal(template))
			Expect(cloudFormationClient.CreateStackCall.Receives.Input.Parameters).To(ConsistOf(
				[]*cloudformation.Parameter{
					&cloudformation.Parameter{ParameterKey: aws.String("a"), ParameterValue: aws.String("b")},
					&cloudformation.Parameter{ParameterKey: aws.String("c"), ParameterValue: aws.String("d")},
					&cloudformation.Parameter{ParameterKey: aws.String("e"), ParameterValue: aws.String("f")},
				}))
			Expect(cloudFormationClient.CreateStackCall.Receives.Input.Tags).To(ConsistOf(
				[]*cloudformation.Tag{
					&cloudformation.Tag{Key: aws.String("Name"), Value: aws.String(stackName)},
				}))
			Expect(cloudFormationClient.CreateStackCall.Receives.Input.Capabilities).To(Equal([]*string{aws.String("CAPABILITY_IAM")}))

			Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(BeNil())
		})

		Context("when creating the stack fails", func() {
			It("should return the error", func() {
				theError := awserr.New("SomeCode", "some message", nil)
				cloudFormationClient.CreateStackCall.Returns.Error = theError

				Expect(client.UpsertStack(stackName, template, parameters)).To(MatchError(theError))
			})
		})
	})

	Context("when the stack exists", func() {
		Context("when the stack has never been updated", func() {
			BeforeEach(func() {
				cloudFormationClient.DescribeStacksCall.Returns.Output.Stacks[0].StackStatus = aws.String("CREATE_COMPLETE")
			})

			It("should try to update the stack, not create it", func() {
				Expect(client.UpsertStack(stackName, template, parameters)).To(Succeed())

				Expect(cloudFormationClient.CreateStackCall.Receives.Input).To(BeNil())

				Expect(*cloudFormationClient.UpdateStackCall.Receives.Input.StackName).To(Equal(stackName))
				Expect(*cloudFormationClient.UpdateStackCall.Receives.Input.TemplateBody).To(Equal(template))
				Expect(cloudFormationClient.UpdateStackCall.Receives.Input.Parameters).To(ConsistOf(
					[]*cloudformation.Parameter{
						&cloudformation.Parameter{ParameterKey: aws.String("a"), ParameterValue: aws.String("b")},
						&cloudformation.Parameter{ParameterKey: aws.String("c"), ParameterValue: aws.String("d")},
						&cloudformation.Parameter{ParameterKey: aws.String("e"), ParameterValue: aws.String("f")},
					}))
				Expect(cloudFormationClient.UpdateStackCall.Receives.Input.Capabilities).To(Equal([]*string{aws.String("CAPABILITY_IAM")}))
			})

			Context("when UpdateStack returns an error because there are no changes", func() {
				BeforeEach(func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(
						awserr.New("ValidationError", "No updates are to be performed.", nil),
						400, "some-request-id")
				})

				It("should swallow the error and succeed", func() {
					Expect(client.UpsertStack(stackName, template, parameters)).To(Succeed())
				})
			})

			Context("when UpdateStack errors for a different reason", func() {
				It("should return the error", func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = errors.New("some error")

					Expect(client.UpsertStack(stackName, template, parameters)).To(MatchError("some error"))
				})
			})
		})

		Context("when the stack has been updated in the past", func() {
			BeforeEach(func() {
				cloudFormationClient.DescribeStacksCall.Returns.Output.Stacks[0].StackStatus = aws.String("UPDATE_COMPLETE")
			})

			It("should try to update the stack, not create it", func() {
				Expect(client.UpsertStack(stackName, template, parameters)).To(Succeed())

				Expect(cloudFormationClient.CreateStackCall.Receives.Input).To(BeNil())

				Expect(*cloudFormationClient.UpdateStackCall.Receives.Input.StackName).To(Equal(stackName))
				Expect(*cloudFormationClient.UpdateStackCall.Receives.Input.TemplateBody).To(Equal(template))
				Expect(cloudFormationClient.UpdateStackCall.Receives.Input.Parameters).To(ConsistOf(
					[]*cloudformation.Parameter{
						&cloudformation.Parameter{ParameterKey: aws.String("a"), ParameterValue: aws.String("b")},
						&cloudformation.Parameter{ParameterKey: aws.String("c"), ParameterValue: aws.String("d")},
						&cloudformation.Parameter{ParameterKey: aws.String("e"), ParameterValue: aws.String("f")},
					}))
			})

			Context("when UpdateStack returns an error because there are no changes", func() {
				BeforeEach(func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(
						awserr.New("ValidationError", "No updates are to be performed.", nil),
						400, "some-request-id")
				})

				It("should swallow the error and succeed", func() {
					Expect(client.UpsertStack(stackName, template, parameters)).To(Succeed())
				})
			})

			Context("when UpdateStack errors for a different reason", func() {
				It("should return the error", func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = errors.New("some error")

					Expect(client.UpsertStack(stackName, template, parameters)).To(MatchError("some error"))
				})
			})
		})

		Context("when the stack exists but is in an unhealthy state", func() {
			BeforeEach(func() {
				cloudFormationClient.DescribeStacksCall.Returns.Output.Stacks[0].StackStatus = aws.String("SOME_UNHEALTHY_STATE")
			})

			It("should return an error", func() {
				err := client.UpsertStack(stackName, template, parameters)

				Expect(err).To(MatchError(fmt.Sprintf("refusing to update stack %q, status %q",
					stackName, "SOME_UNHEALTHY_STATE")))
			})

			It("should not attempt to create or update the stack", func() {
				client.UpsertStack(stackName, template, parameters)

				Expect(cloudFormationClient.CreateStackCall.Receives.Input).To(BeNil())
				Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(BeNil())
			})
		})
	})

	Context("when describing the stack fails for an unknown reason", func() {
		BeforeEach(func() {
			cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(
				awserr.New("ValidationError", "Bad input data!", nil), 400, "some-request-id")
		})
		It("should return the error", func() {
			Expect(client.UpsertStack(stackName, template, parameters)).To(MatchError(ContainSubstring("Bad input data!")))
		})
		It("should not attempt to create or update the stack", func() {
			client.UpsertStack(stackName, template, parameters)

			Expect(cloudFormationClient.CreateStackCall.Receives.Input).To(BeNil())
			Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(BeNil())
		})
	})
})
