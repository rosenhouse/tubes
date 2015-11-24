package matchers

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/rosenhouse/awsfaker"
)

func requestFailure_to_errorResponse(e awserr.RequestFailure) *awsfaker.ErrorResponse {
	if e == nil {
		return nil
	}
	return &awsfaker.ErrorResponse{
		HTTPStatusCode:  e.StatusCode(),
		AWSErrorCode:    e.Code(),
		AWSErrorMessage: e.Message(),
	}
}

func MatchErrorResponse(errRespToMatch *awsfaker.ErrorResponse) types.GomegaMatcher {
	return gomega.WithTransform(requestFailure_to_errorResponse, gomega.Equal(errRespToMatch))
}
