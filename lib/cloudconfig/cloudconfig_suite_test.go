package cloudconfig_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCloudconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudconfig Suite")
}
