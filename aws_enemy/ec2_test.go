package aws_enemy_test

import (
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EC2 Key Pairs", func() {
	var keyName string

	BeforeEach(func() {
		keyName = fmt.Sprintf("test-%x", rand.Int())
	})

	AfterEach(func() {
		_, err := ec2Client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
			KeyName: aws.String(keyName),
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("DeleteKeyPair", func() {
		Context("when the keypair does not exist", func() {
			It("returns a ValidationError", func() {
				_, err := ec2Client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
					KeyName: aws.String(keyName),
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CreateKeyPair", func() {
		Context("when the keypair already exists", func() {
			It("returns an InvalidKeyPair.Duplicate error", func() {
				output, err := ec2Client.CreateKeyPair(&ec2.CreateKeyPairInput{
					KeyName: aws.String(keyName),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				Expect(*output.KeyName).To(Equal(keyName))

				output, err = ec2Client.CreateKeyPair(&ec2.CreateKeyPairInput{
					KeyName: aws.String(keyName),
				})
				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.StatusCode()).To(Equal(400))
				Expect(awsErr.Code()).To(Equal("InvalidKeyPair.Duplicate"))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("The keypair '%s' already exists.", keyName)))
			})
		})
	})
})
