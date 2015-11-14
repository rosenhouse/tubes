package aws_enemy_test

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/rosenhouse/tubes/lib/awsclient"

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
		client    *awsclient.Client
		stackName string
	)

	BeforeEach(func() {
		client = awsclient.New(awsclient.Config{
			Region:    os.Getenv("AWS_DEFAULT_REGION"),
			AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		})
		stackName = fmt.Sprintf("test-stack-%x", rand.Int63())
	})
	AfterEach(func() {
		client.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	Describe("UpdateStack", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError", func() {
				_, err := client.CloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.Code()).To(Equal("ValidationError"))
				Expect(awsErr.StatusCode()).To(Equal(400))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("Stack [%s] does not exist", stackName)))
			})
		})
		Context("when the stack exists but there are no changes", func() {
			BeforeEach(func() {
				_, err := client.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() (string, error) {
					output, err := client.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
						StackName: aws.String(stackName),
					})
					if err != nil {
						return "", err
					}
					return *output.Stacks[0].StackStatus, nil
				}, "60s", "5s").Should(Equal("CREATE_COMPLETE"))
			})
			It("returns a ValidationError", func() {
				_, err := client.CloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.Code()).To(Equal("ValidationError"))
				Expect(awsErr.Message()).To(Equal("No updates are to be performed."))
				Expect(awsErr.StatusCode()).To(Equal(400))
			})
		})
	})

	Describe("CreateStack", func() {
		Context("when the stack already exists", func() {
			It("returns an AlreadyExists error", func() {
				_, err := client.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = client.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.Code()).To(Equal("AlreadyExistsException"))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("Stack [%s] already exists", stackName)))
				Expect(awsErr.StatusCode()).To(Equal(400))
			})
		})
	})

	Describe("DescribeStacks", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError error", func() {
				_, err := client.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
					StackName: aws.String(stackName),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.Code()).To(Equal("ValidationError"))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("Stack with id %s does not exist", stackName)))
				Expect(awsErr.StatusCode()).To(Equal(400))
			})
		})
	})

	Describe("DescribeStackResources", func() {
		Context("when the stack does not exist", func() {
			It("returns a ValidationError error", func() {
				_, err := client.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
					StackName: aws.String(stackName),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.Code()).To(Equal("ValidationError"))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("Stack with id %s does not exist", stackName)))
				Expect(awsErr.StatusCode()).To(Equal(400))
			})
		})
	})

	Describe("DeleteStack", func() {
		Context("when the stack does not exist", func() {
			It("succeeds", func() {
				_, err := client.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
					StackName: aws.String(stackName),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
