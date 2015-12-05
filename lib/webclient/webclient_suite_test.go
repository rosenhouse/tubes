package webclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWebclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webclient Suite")
}
