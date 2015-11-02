package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/rosenhouse/tubes/lib/matchers"
)

var _ = Describe("MatchYAMLMatcher", func() {
	Context("When passed stringifiables", func() {
		Describe("parsing pure JSON", func() {
			It("should succeed if the YAML matches", func() {
				Ω("{}").Should(MatchYAML("{}"))
				Ω(`{"a":1}`).Should(MatchYAML(`{"a":1}`))
				Ω(`{
			             "a":1
			         }`).Should(MatchYAML(`{"a":1}`))
				Ω(`{"a":1, "b":2}`).Should(MatchYAML(`{"b":2, "a":1}`))
				Ω(`{"a":1}`).ShouldNot(MatchYAML(`{"b":2, "a":1}`))
			})

			It("should work with byte arrays", func() {
				Ω([]byte("{}")).Should(MatchYAML([]byte("{}")))
				Ω("{}").Should(MatchYAML([]byte("{}")))
				Ω([]byte("{}")).Should(MatchYAML("{}"))
			})
		})

		Describe("parsing non-JSON YAML", func() {
			It("should succeed if the YAML matches", func() {
				Ω("").Should(MatchYAML(""))
				Ω(`a: 1`).Should(MatchYAML(`a:    1`))
				Ω(`
---
a: 1  # <-- some comment
b: 2
d:
  - one
  - two
`).Should(MatchYAML(`{"b":2, "a":1, "d": ["one","two"]}`))
				Ω(`a:1}`).ShouldNot(MatchYAML(`{"b":2, "a":1}`))
			})

			It("should correctly handle block literals", func() {
				Ω(`
c: |
   some
   multi
   line
   data
`).Should(MatchYAML(`{"c":"some\nmulti\nline\ndata\n"}`))
			})
		})
	})

	Context("when either side is not valid YAML", func() {
		It("should error", func() {
			success, err := (&MatchYAMLMatcher{YAMLToMatch: `a: :::`}).Match(`{}`)
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())

			success, err = (&MatchYAMLMatcher{YAMLToMatch: `{}`}).Match(`a: :::`)
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("when either side is neither a string nor a stringer", func() {
		It("should error", func() {
			success, err := (&MatchYAMLMatcher{YAMLToMatch: "{}"}).Match(2)
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())

			success, err = (&MatchYAMLMatcher{YAMLToMatch: 2}).Match("{}")
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())

			success, err = (&MatchYAMLMatcher{YAMLToMatch: nil}).Match("{}")
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())

			success, err = (&MatchYAMLMatcher{YAMLToMatch: 2}).Match(nil)
			Ω(success).Should(BeFalse())
			Ω(err).Should(HaveOccurred())
		})
	})
})
