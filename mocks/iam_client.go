package mocks

import "github.com/aws/aws-sdk-go/service/iam"

type IAMClient struct {
	DeleteUserCall struct {
		Receives struct {
			Input *iam.DeleteUserInput
		}
		Returns struct {
			Output *iam.DeleteUserOutput
			Error  error
		}
	}

	CreateAccessKeyCall struct {
		Receives struct {
			Input *iam.CreateAccessKeyInput
		}
		Returns struct {
			Output *iam.CreateAccessKeyOutput
			Error  error
		}
	}

	DeleteAccessKeyCall struct {
		Receives struct {
			Input *iam.DeleteAccessKeyInput
		}
		Returns struct {
			Output *iam.DeleteAccessKeyOutput
			Error  error
		}
	}

	ListAccessKeysCall struct {
		Receives struct {
			Input *iam.ListAccessKeysInput
		}
		Returns struct {
			Output *iam.ListAccessKeysOutput
			Error  error
		}
	}
}

func (c *IAMClient) DeleteUser(input *iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	c.DeleteUserCall.Receives.Input = input
	return c.DeleteUserCall.Returns.Output, c.DeleteUserCall.Returns.Error
}

func (c *IAMClient) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	c.CreateAccessKeyCall.Receives.Input = input
	return c.CreateAccessKeyCall.Returns.Output, c.CreateAccessKeyCall.Returns.Error
}

func (c *IAMClient) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	c.DeleteAccessKeyCall.Receives.Input = input
	return c.DeleteAccessKeyCall.Returns.Output, c.DeleteAccessKeyCall.Returns.Error
}

func (c *IAMClient) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	c.ListAccessKeysCall.Receives.Input = input
	return c.ListAccessKeysCall.Returns.Output, c.ListAccessKeysCall.Returns.Error
}
