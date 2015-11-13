package aws_enemy_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/rosenhouse/tubes/lib/awsclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mocking out an AWS service over the network", func() {
	var (
		client    *awsclient.Client
		stackName string
		server    *httptest.Server
		handler   func(w http.ResponseWriter, r *http.Request)
	)

	BeforeEach(func() {
		handler = nil
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if handler == nil {
				panic("test was not properly set up: missing handler")
			}
			handler(w, r)
		}))

		client = awsclient.New(awsclient.Config{
			Region:           "some-region",
			AccessKey:        "some-access-key",
			SecretKey:        "some-secret-key",
			EndpointOverride: server.URL,
		})

		stackName = fmt.Sprintf("test-stack-%x", rand.Int63())
	})

	AfterEach(func() {
		server.Close()
	})

	writeResponse := func(w http.ResponseWriter, statusCode int, data interface{}) {
		responseBodyBytes, err := xml.Marshal(data)
		Expect(err).NotTo(HaveOccurred())

		w.WriteHeader(statusCode)
		_, err = w.Write(responseBodyBytes)
		Expect(err).NotTo(HaveOccurred())
	}

	type ErrorResponse struct {
		XMLName   xml.Name `xml:"ErrorResponse"`
		Code      string   `xml:"Error>Code"`
		Message   string   `xml:"Error>Message"`
		RequestID string   `xml:"RequestId"`
	}

	Describe("UpdateStack", func() {
		parseRequest := func(r *http.Request) cloudformation.UpdateStackInput {
			requestBodyBytes, err := ioutil.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			values, err := url.ParseQuery(string(requestBodyBytes))
			Expect(err).NotTo(HaveOccurred())
			Expect(values["Action"]).To(Equal([]string{"UpdateStack"}))

			return cloudformation.UpdateStackInput{
				StackName: aws.String(values.Get("StackName")),
			}
		}

		Context("when the stack exists and ready for updates", func() {
			BeforeEach(func() {
				handler = func(w http.ResponseWriter, r *http.Request) {
					stackName := *parseRequest(r).StackName
					writeResponse(w, http.StatusOK, cloudformation.UpdateStackOutput{
						StackId: aws.String("some-stack-id-for-" + stackName),
					})
				}
			})

			It("succeeds", func() {
				output, err := client.CloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				Expect(*output.StackId).To(Equal("some-stack-id-for-" + stackName))
			})
		})

		Context("when the stack does not exist", func() {
			BeforeEach(func() {
				handler = func(w http.ResponseWriter, r *http.Request) {
					stackName := *parseRequest(r).StackName
					writeResponse(w, http.StatusBadRequest, ErrorResponse{
						Code:    "ValidationError",
						Message: fmt.Sprintf("Stack [%s] does not exist", stackName),
					})
				}
			})

			It("returns a ValidationError", func() {
				_, err := client.CloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
					StackName:    aws.String(stackName),
					TemplateBody: aws.String(templateBody),
				})

				Expect(err).To(HaveOccurred())
				awsErr := err.(awserr.RequestFailure)
				Expect(awsErr.StatusCode()).To(Equal(400))
				Expect(awsErr.Code()).To(Equal("ValidationError"))
				Expect(awsErr.Message()).To(Equal(fmt.Sprintf("Stack [%s] does not exist", stackName)))
			})
		})
	})
})
