package director_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDirector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Director Suite")
}
