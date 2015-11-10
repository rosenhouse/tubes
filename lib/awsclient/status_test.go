package awsclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("status reports", func() {
	var pundit awsclient.CloudFormationStatusPundit
	BeforeEach(func() { pundit = awsclient.CloudFormationStatusPundit{} })

	It("reports the healthy statuses as such", func() {
		for _, statusString := range []string{
			"CREATE_IN_PROGRESS",
			"CREATE_COMPLETE",
			"UPDATE_IN_PROGRESS",
			"UPDATE_COMPLETE_CLEANUP_IN_PROGRESS",
			"UPDATE_COMPLETE",
		} {
			Expect(pundit.IsHealthy(statusString)).To(BeTrue())
		}
	})
	It("reports the unhealthy statuses as such", func() {
		for _, statusString := range []string{
			"CREATE_FAILED",
			"ROLLBACK_IN_PROGRESS",
			"ROLLBACK_FAILED",
			"ROLLBACK_COMPLETE",
			"DELETE_IN_PROGRESS",
			"DELETE_FAILED",
			"DELETE_COMPLETE",
			"UPDATE_ROLLBACK_IN_PROGRESS",
			"UPDATE_ROLLBACK_FAILED",
			"UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS",
			"UPDATE_ROLLBACK_COMPLETE",
		} {
			Expect(pundit.IsHealthy(statusString)).To(BeFalse())
		}
	})
	It("reports the complete statuses as such", func() {
		for _, statusString := range []string{
			"CREATE_COMPLETE",
			"ROLLBACK_COMPLETE",
			"DELETE_COMPLETE",
			"UPDATE_COMPLETE",
			"UPDATE_ROLLBACK_COMPLETE",
		} {
			Expect(pundit.IsComplete(statusString)).To(BeTrue())
		}
	})
	It("reports the incomplete statuses as such", func() {
		for _, statusString := range []string{
			"CREATE_IN_PROGRESS",
			"CREATE_FAILED",
			"ROLLBACK_IN_PROGRESS",
			"ROLLBACK_FAILED",
			"DELETE_IN_PROGRESS",
			"DELETE_FAILED",
			"UPDATE_IN_PROGRESS",
			"UPDATE_COMPLETE_CLEANUP_IN_PROGRESS",
			"UPDATE_ROLLBACK_IN_PROGRESS",
			"UPDATE_ROLLBACK_FAILED",
			"UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS",
		} {
			Expect(pundit.IsComplete(statusString)).To(BeFalse())
		}
	})
})
