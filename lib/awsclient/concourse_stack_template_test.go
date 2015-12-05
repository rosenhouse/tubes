package awsclient_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("Generating the concourse template", func() {
	It("should match the fixture", func() {
		asJSON := awsclient.ConcourseStackTemplate.String()

		expected, err := ioutil.ReadFile("fixtures/concourse_stack_template.json")
		Expect(err).NotTo(HaveOccurred())

		Expect(asJSON).To(MatchJSON(expected))
	})
})
