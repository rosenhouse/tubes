package awsclient_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("Generating the base template", func() {
	It("should match the fixture", func() {
		asJSON := awsclient.BaseStackTemplate.String()

		expected, err := ioutil.ReadFile("fixtures/base_stack_template.json")
		Expect(err).NotTo(HaveOccurred())

		Expect(asJSON).To(MatchJSON(expected))
	})
})
