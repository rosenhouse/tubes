package awsclient_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/mocks"
)

var _ = Describe("User operations", func() {
	var (
		client    awsclient.Client
		iamClient *mocks.IAMClient
		userName  string
	)

	BeforeEach(func() {
		iamClient = &mocks.IAMClient{}
		client = awsclient.Client{
			IAM: iamClient,
		}
		userName = fmt.Sprintf("some-user-%x", rand.Int31()>>16)
	})

	Describe("CreateAccessKey", func() {
		It("should call the SDK CreateAccessKey function", func() {
			iamClient.CreateAccessKeyCall.Returns.Output = &iam.CreateAccessKeyOutput{
				AccessKey: &iam.AccessKey{
					AccessKeyId:     aws.String("some-access-key"),
					SecretAccessKey: aws.String("some-secret-access-key"),
				},
			}

			accessKey, secretKey, err := client.CreateAccessKey(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(accessKey).To(Equal("some-access-key"))
			Expect(secretKey).To(Equal("some-secret-access-key"))

			Expect(*iamClient.CreateAccessKeyCall.Receives.Input.UserName).To(Equal(userName))
		})

		Context("when the SDK returns an error", func() {
			It("should return the error", func() {
				iamClient.CreateAccessKeyCall.Returns.Error = errors.New("some error")

				_, _, err := client.CreateAccessKey(userName)
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("DeleteAccessKey", func() {
		It("should call the SDK DeleteAccessKey function", func() {
			Expect(client.DeleteAccessKey(userName, "some-access-key")).To(Succeed())

			Expect(*iamClient.DeleteAccessKeyCall.Receives.Input.UserName).To(Equal(userName))
			Expect(*iamClient.DeleteAccessKeyCall.Receives.Input.AccessKeyId).To(Equal("some-access-key"))
		})

		Context("when the SDK returns an error", func() {
			It("should return the error", func() {
				iamClient.DeleteAccessKeyCall.Returns.Error = errors.New("some error")

				Expect(client.DeleteAccessKey(userName, "some-access-key")).To(MatchError("some error"))
			})
		})
	})

	Describe("ListAccessKeys", func() {
		It("should call the SDK ListAccessKeys function", func() {
			iamClient.ListAccessKeysCall.Returns.Output = &iam.ListAccessKeysOutput{
				AccessKeyMetadata: []*iam.AccessKeyMetadata{
					&iam.AccessKeyMetadata{
						AccessKeyId: aws.String("one"),
					},
					&iam.AccessKeyMetadata{
						AccessKeyId: aws.String("two"),
					},
				},
			}
			keys, err := client.ListAccessKeys(userName)
			Expect(err).NotTo(HaveOccurred())

			Expect(*iamClient.ListAccessKeysCall.Receives.Input.UserName).To(Equal(userName))
			Expect(keys).To(Equal([]string{"one", "two"}))
		})

		Context("when the SDK returns an error", func() {
			It("should return the error", func() {
				iamClient.ListAccessKeysCall.Returns.Error = errors.New("some error")

				_, err := client.ListAccessKeys(userName)
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
