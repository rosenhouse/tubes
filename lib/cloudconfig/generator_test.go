package cloudconfig_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/cloudconfig"
)

var _ = Describe("Generator", func() {

	var (
		stackResources map[string]string
		generator      *cloudconfig.Generator
	)

	BeforeEach(func() {
		stackResources = map[string]string{
			"ConcourseSecurityGroup": "some-concourse-security-group-id",
			"ConcourseSubnet":        "some-concourse-subnet-id",
			"LoadBalancer":           "some-concourse-elb",
		}

		generator = &cloudconfig.Generator{}
	})

	XIt("generates a cloud config file using the given stack resources", func() {
		cloudConfigBytes, err := generator.Generate(stackResources)
		Expect(err).NotTo(HaveOccurred())
		Expect(cloudConfigBytes).To(Equal("some magic yaml"))
	})
})
