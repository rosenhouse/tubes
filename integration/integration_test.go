package integration_test

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/awsclient"
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

var _ = Describe("Integration", func() {
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
		stackName = fmt.Sprintf("test-stack-%d", rand.Int63())
	})
	AfterEach(func() {
		client.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
	})

	Describe("CloudFormation", func() {
		Describe("UpdateStack", func() {
			Context("when the stack does not exist", func() {
				It("should succeed", func() {
					_, err := client.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
						StackName:    aws.String(stackName),
						TemplateBody: aws.String(templateBody),
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
		Describe("CreateStack", func() {
			Context("when the stack already exists", func() {
				It("should return an AlreadyExists error", func() {
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
					awserr := err.(awserr.Error)
					Expect(awserr.Code()).To(Equal("AlreadyExistsException"))
				})
			})
		})
		Describe("DescribeStacks", func() {
			Context("when the stack does not exist", func() {
				It("should return a ValidationError error", func() {
					_, err := client.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
						StackName: aws.String(stackName),
					})
					Expect(err).To(HaveOccurred())
					awsErr := err.(awserr.Error)
					Expect(awsErr.Code()).To(Equal("ValidationError"))
					Expect(awsErr.Message()).To(ContainSubstring("does not exist"))
				})
			})
		})
	})
})
