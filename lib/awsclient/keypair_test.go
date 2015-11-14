package awsclient_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("Keypair operations", func() {
	var (
		client    awsclient.Client
		ec2Client *mocks.EC2Client
		keyName   string
	)

	BeforeEach(func() {
		ec2Client = &mocks.EC2Client{}
		client = awsclient.Client{
			EC2: ec2Client,
		}
		keyName = fmt.Sprintf("some-key-%x", rand.Int31()>>16)
	})

	Describe("CreateKeyPair", func() {
		It("should call the SDK CreateKeyPair function and return the resulting private key", func() {
			ec2Client.CreateKeyPairCall.Returns.Output = &ec2.CreateKeyPairOutput{}
			ec2Client.CreateKeyPairCall.Returns.Output.KeyMaterial = aws.String("some pem block")
			key, err := client.CreateKeyPair(keyName)
			Expect(err).NotTo(HaveOccurred())

			Expect(key).To(Equal("some pem block"))

			Expect(*ec2Client.CreateKeyPairCall.Receives.Input.KeyName).To(Equal(keyName))
		})

		Context("when the SDK returns an error", func() {
			It("should return the error", func() {
				ec2Client.CreateKeyPairCall.Returns.Error = errors.New("some error")

				_, err := client.CreateKeyPair(keyName)
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("DeleteKeyPair", func() {
		It("should call the SDK DeleteKeyPair function", func() {
			Expect(client.DeleteKeyPair(keyName)).To(Succeed())

			Expect(*ec2Client.DeleteKeyPairCall.Receives.Input.KeyName).To(Equal(keyName))
		})

		Context("when the SDK returns an error", func() {
			It("should return the error", func() {
				ec2Client.DeleteKeyPairCall.Returns.Error = errors.New("some error")

				Expect(client.DeleteKeyPair(keyName)).To(MatchError("some error"))
			})
		})
	})
})
