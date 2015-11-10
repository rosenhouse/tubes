package awsclient

type CloudFormationStatusPundit struct{}

func (p CloudFormationStatusPundit) IsHealthy(statusString string) bool {
	switch statusString {
	case "CREATE_IN_PROGRESS", "CREATE_COMPLETE", "UPDATE_IN_PROGRESS", "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS", "UPDATE_COMPLETE":
		return true
	}
	return false
}

func (p CloudFormationStatusPundit) IsComplete(statusString string) bool {
	switch statusString {
	case "CREATE_COMPLETE", "ROLLBACK_COMPLETE", "DELETE_COMPLETE", "UPDATE_COMPLETE", "UPDATE_ROLLBACK_COMPLETE":
		return true
	}
	return false
}
