package credentials_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/lib/credentials"
)

var _ = Describe("Generating credentials", func() {
	It("should populate a struct's public fields", func() {
		var myCreds struct {
			BasicAuthPassword string
			DBPassword        string
		}

		generator := credentials.Generator{Length: 15}
		Expect(generator.Fill(&myCreds)).To(Succeed())
		Expect(myCreds.BasicAuthPassword).To(HaveLen(15))
		Expect(myCreds.DBPassword).To(HaveLen(15))
	})

	It("should set different strings every time", func() {
		myCredsList := make([]struct {
			BasicAuthPassword string
			DBPassword        string
		}, 7)

		generator := credentials.Generator{Length: 15}

		// sample the generator, collect into a histogram
		allCreds := map[string]int{}
		for i := 0; i < 7; i++ {
			Expect(generator.Fill(&myCredsList[i])).To(Succeed())
			allCreds[myCredsList[i].BasicAuthPassword]++
			allCreds[myCredsList[i].DBPassword]++
		}

		// check that no password appeared more than once
		Expect(allCreds).To(HaveLen(7 * 2))
		for _, count := range allCreds {
			Expect(count).To(Equal(1))
		}
	})

	It("should work with any requested length", func() {
		for length := 1; length < 45; length++ {
			generator := credentials.Generator{Length: length}
			var creds struct{ Value string }
			Expect(generator.Fill(&creds)).To(Succeed())
			Expect(creds.Value).To(HaveLen(length))
		}
	})

	Context("when provided an invalid type", func() {
		It("should return an informative error", func() {
			generator := credentials.Generator{Length: 7}
			var creds struct{ Value string }
			Expect(generator.Fill(creds)).To(MatchError("expecting a pointer to a struct"))

			var aString string
			Expect(generator.Fill(&aString)).To(MatchError("expecting a pointer to a struct"))

			var nope *struct{ Value string }
			Expect(generator.Fill(nope)).To(MatchError("pointer must not be nil"))
		})
	})

	Context("when the length is invalid", func() {
		It("should return an informative error", func() {
			generator := credentials.Generator{Length: 0}
			var creds struct{ Value string }
			Expect(generator.Fill(&creds)).To(MatchError("length must be positive"))

			generator = credentials.Generator{Length: -7}
			Expect(generator.Fill(&creds)).To(MatchError("length must be positive"))
		})
	})
})
