package aws_enemy_test

import (
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/rosenhouse/tubes/aws_enemy"

	. "github.com/rosenhouse/tubes/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const templateBody = `{
"AWSTemplateFormatVersion": "2010-09-09",
"Resources": {
  "NATSecurityGroup": {
    "Type": "AWS::EC2::SecurityGroup",
    "Properties": {
      "SecurityGroupIngress": [
        {
          "ToPort": "22",
          "IpProtocol": "tcp",
          "FromPort": "22",
          "CidrIp": "0.0.0.0/0"
        }
      ],
      "GroupDescription": "test-group",
      "SecurityGroupEgress": []
    }
  }
}
}`

var _ = Describe("CloudFormation error cases", func() {
	var (
		stackName string
		cfErrors  aws_enemy.CloudFormation
	)

	BeforeEach(func() {
		stackName = fmt.Sprintf("test-stack-%x", rand.Int63())
	})

	AfterEach(func() {
		cloudformationClient.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	Describe("UpdateStack", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError", func() {
				_, err := cloudformationClient.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				expectedErrorResp := cfErrors.UpdateStack_StackMissingError(stackName)
				Expect(err).To(MatchErrorResponse(expectedErrorResp))
			})
		})
		Context("when the stack exists but there are no changes", func() {
			BeforeEach(func() {
				_, err := cloudformationClient.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() (string, error) {
					output, err := cloudformationClient.DescribeStacks(&cloudformation.DescribeStacksInput{
						StackName: aws.String(stackName),
					})
					if err != nil {
						return "", err
					}
					return *output.Stacks[0].StackStatus, nil
				}, "60s", "5s").Should(Equal("CREATE_COMPLETE"))
			})
			It("returns a ValidationError", func() {
				_, err := cloudformationClient.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				expectedErrorResp := cfErrors.UpdateStack_NoChangesError()
				Expect(err).To(MatchErrorResponse(expectedErrorResp))
			})
		})
	})

	Describe("CreateStack", func() {
		Context("when the stack already exists", func() {
			It("returns an AlreadyExists error", func() {
				_, err := cloudformationClient.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = cloudformationClient.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				expectedErrorResp := cfErrors.CreateStack_AlreadyExistsError(stackName)
				Expect(err).To(MatchErrorResponse(expectedErrorResp))
			})
		})
	})

	Describe("DescribeStacks", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError error", func() {
				_, err := cloudformationClient.DescribeStacks(&cloudformation.DescribeStacksInput{
					StackName: aws.String(stackName),
				})
				Expect(err).To(HaveOccurred())
				expectedErrorResp := cfErrors.DescribeStacks_StackMissingError(stackName)
				Expect(err).To(MatchErrorResponse(expectedErrorResp))
			})
		})
	})

	Describe("DescribeStackResources", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError error", func() {
				_, err := cloudformationClient.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
					StackName: aws.String(stackName),
				})
				Expect(err).To(HaveOccurred())
				expectedErrorResp := cfErrors.DescribeStackResources_StackMissingError(stackName)
				Expect(err).To(MatchErrorResponse(expectedErrorResp))
			})
		})
	})

	Describe("DeleteStack", func() {
		Context("when the stack does not exist", func() {
			It("succeeds", func() {
				_, err := cloudformationClient.DeleteStack(&cloudformation.DeleteStackInput{
					StackName: aws.String(stackName),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
