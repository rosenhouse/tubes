package awsclient

type CloudFormationUpsertPundit struct{}

func (p CloudFormationUpsertPundit) IsHealthy(statusString string) bool {
	switch statusString {
	case "CREATE_IN_PROGRESS", "CREATE_COMPLETE", "UPDATE_IN_PROGRESS", "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS", "UPDATE_COMPLETE":
		return true
	}
	return false
}

func (p CloudFormationUpsertPundit) IsComplete(statusString string) bool {
	switch statusString {
	case "CREATE_COMPLETE", "ROLLBACK_COMPLETE", "DELETE_COMPLETE", "UPDATE_COMPLETE", "UPDATE_ROLLBACK_COMPLETE":
		return true
	}
	return false
}

type CloudFormationDeletePundit struct{}

func (p CloudFormationDeletePundit) IsHealthy(statusString string) bool {
	switch statusString {
	case "DELETE_IN_PROGRESS", "DELETE_COMPLETE":
		return true
	}
	return false
}

func (p CloudFormationDeletePundit) IsComplete(statusString string) bool {
	return CloudFormationUpsertPundit{}.IsComplete(statusString)
}
