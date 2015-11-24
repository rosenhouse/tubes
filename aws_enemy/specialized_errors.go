package aws_enemy

import (
	"fmt"
	"net/http"

	"github.com/rosenhouse/awsfaker"
)

type EC2 struct{}

func (EC2) CreateKeyPair_AlreadyExistsError(keypairName string) *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "InvalidKeyPair.Duplicate",
		AWSErrorMessage: fmt.Sprintf("The keypair '%s' already exists.", keypairName),
	}
}

type CloudFormation struct{}

func (CloudFormation) UpdateStack_StackMissingError(stackName string) *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "ValidationError",
		AWSErrorMessage: fmt.Sprintf("Stack [%s] does not exist", stackName),
	}
}

func (CloudFormation) UpdateStack_NoChangesError() *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "ValidationError",
		AWSErrorMessage: "No updates are to be performed.",
	}
}

func (CloudFormation) CreateStack_AlreadyExistsError(stackName string) *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "AlreadyExistsException",
		AWSErrorMessage: fmt.Sprintf("Stack [%s] already exists", stackName),
	}
}

func (CloudFormation) DescribeStacks_StackMissingError(stackName string) *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "ValidationError",
		AWSErrorMessage: fmt.Sprintf("Stack with id %s does not exist", stackName),
	}
}

func (CloudFormation) DescribeStackResources_StackMissingError(stackName string) *awsfaker.ErrorResponse {
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		AWSErrorCode:    "ValidationError",
		AWSErrorMessage: fmt.Sprintf("Stack with id %s does not exist", stackName),
	}
}
