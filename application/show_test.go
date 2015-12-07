package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/application"
)

type erroringWriter struct{}

func (w *erroringWriter) Write(data []byte) (int, error) {
	return -3, errors.New("write failed")
}

var _ = Describe("Show", func() {

	var options application.ShowOptions

	Context("when all options are empty", func() {
		It("should return a friendly error", func() {
			Expect(app.Show(stackName, application.ShowOptions{})).To(MatchError("set at least one flag"))
		})
	})

	Context("when the SSH key option is set", func() {
		BeforeEach(func() { options.SSHKey = true })

		It("should print the SSH key to the result writer", func() {
			configStore.Values["ssh-key"] = []byte("some pem block")

			Expect(app.Show(stackName, options)).To(Succeed())

			Expect(resultBuffer.Contents()).To(Equal([]byte("some pem block")))
		})

		Context("when the config store get errors", func() {
			It("should return the error", func() {
				configStore.Errors["ssh-key"] = errors.New("some error")
				Expect(app.Show(stackName, options)).To(MatchError("some error"))
			})
		})
	})

	Context("when the BOSH IP option is set", func() {
		BeforeEach(func() { options.BoshIP = true })

		It("should print the BOSH IP to the result writer", func() {
			configStore.Values["bosh-ip"] = []byte("some ip address")

			Expect(app.Show(stackName, options)).To(Succeed())

			Expect(resultBuffer.Contents()).To(Equal([]byte("some ip address")))
		})

		Context("when the config store get errors", func() {
			It("should return the error", func() {
				configStore.Errors["bosh-ip"] = errors.New("some error")
				Expect(app.Show(stackName, options)).To(MatchError("some error"))
			})
		})
	})

	Context("when the BOSH password option is set", func() {
		BeforeEach(func() { options.BoshPassword = true })

		It("should print the BOSH password to the result writer", func() {
			configStore.Values["bosh-password"] = []byte("some-password")

			Expect(app.Show(stackName, options)).To(Succeed())

			Expect(resultBuffer.Contents()).To(Equal([]byte("some-password")))
		})

		Context("when the config store get errors", func() {
			It("should return the error", func() {
				configStore.Errors["bosh-password"] = errors.New("some error")
				Expect(app.Show(stackName, options)).To(MatchError("some error"))
			})
		})
	})

	Context("when writing the result errors", func() {
		It("should return the error", func() {
			options.SSHKey = true
			app.ResultWriter = &erroringWriter{}
			Expect(app.Show(stackName, options)).To(MatchError("write failed"))
		})
	})
})
