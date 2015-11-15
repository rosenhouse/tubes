package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type erroringWriter struct{}

func (w *erroringWriter) Write(data []byte) (int, error) {
	return -3, errors.New("write failed")
}

var _ = Describe("Show", func() {
	It("should print the SSH key to the result writer", func() {
		configStore.GetCall.Returns.Value = []byte("some pem block")

		Expect(app.Show(stackName)).To(Succeed())

		Expect(resultBuffer.Contents()).To(Equal([]byte("some pem block")))
	})

	It("should construct a key from the stack name", func() {
		Expect(app.Show(stackName)).To(Succeed())
		Expect(configStore.GetCall.Receives.Key).To(Equal(stackName + "/ssh-key"))
	})

	Context("when the config store get errors", func() {
		It("should return the error", func() {
			configStore.GetCall.Returns.Error = errors.New("some error")
			Expect(app.Show(stackName)).To(MatchError("some error"))
		})
	})

	Context("when the writing the result", func() {
		It("should return the error", func() {
			app.ResultWriter = &erroringWriter{}
			Expect(app.Show(stackName)).To(MatchError("write failed"))
		})
	})
})
