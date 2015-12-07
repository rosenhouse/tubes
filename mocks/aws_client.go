package mocks

import "github.com/rosenhouse/tubes/lib/awsclient"

type UpsertStackCall struct {
	Receives struct {
		StackName  string
		Template   string
		Parameters map[string]string
	}
	Returns struct {
		Error error
	}
}

type WaitForStackCall struct {
	Receives struct {
		StackName string
		Pundit    awsclient.CloudFormationStatusPundit
	}
	Returns struct {
		Error error
	}
}

type GetStackResourcesCall struct {
	Receives struct {
		StackName string
	}
	Returns struct {
		Resources map[string]string
		Error     error
	}
}
type DeleteStackCall struct {
	Receives struct {
		StackName string
	}
	Returns struct {
		Error error
	}
}

type AWSClient struct {
	GetLatestNATBoxAMIIDCall struct {
		Returns struct {
			AMIID string
			Error error
		}
	}

	UpsertStackCalls     []UpsertStackCall
	UpsertStackCallCount int

	DeleteStackCalls     []DeleteStackCall
	DeleteStackCallCount int

	WaitForStackCalls     []WaitForStackCall
	WaitForStackCallCount int

	DeleteKeyPairCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Error error
		}
	}
	CreateKeyPairCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			KeyPair string
			Error   error
		}
	}
	GetBaseStackResourcesCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Resources awsclient.BaseStackResources
			Error     error
		}
	}

	GetStackResourcesCalls     []GetStackResourcesCall
	GetStackResourcesCallCount int

	CreateAccessKeyCall struct {
		Receives struct {
			UserName string
		}
		Returns struct {
			AccessKey string
			SecretKey string
			Error     error
		}
	}
	DeleteAccessKeyCall struct {
		Receives struct {
			UserName  string
			AccessKey string
		}
		Returns struct {
			Error error
		}
	}
	ListAccessKeysCall struct {
		Receives struct {
			UserName string
		}
		Returns struct {
			AccessKeys []string
			Error      error
		}
	}
}

func (c *AWSClient) GetLatestNATBoxAMIID() (string, error) {
	return c.GetLatestNATBoxAMIIDCall.Returns.AMIID, c.GetLatestNATBoxAMIIDCall.Returns.Error
}

func (c *AWSClient) UpsertStack(stackName string, template string, parameters map[string]string) error {
	i := c.UpsertStackCallCount
	c.UpsertStackCallCount++

	if i >= len(c.UpsertStackCalls) {
		call := UpsertStackCall{}
		call.Receives.StackName = stackName
		call.Receives.Template = template
		call.Receives.Parameters = parameters
		c.UpsertStackCalls = append(c.UpsertStackCalls, call)
		return nil
	} else {
		c.UpsertStackCalls[i].Receives.StackName = stackName
		c.UpsertStackCalls[i].Receives.Template = template
		c.UpsertStackCalls[i].Receives.Parameters = parameters
		return c.UpsertStackCalls[i].Returns.Error
	}
}

func (c *AWSClient) WaitForStack(stackName string, pundit awsclient.CloudFormationStatusPundit) error {
	i := c.WaitForStackCallCount
	c.WaitForStackCallCount++

	if i >= len(c.WaitForStackCalls) {
		call := WaitForStackCall{}
		call.Receives.StackName = stackName
		call.Receives.Pundit = pundit
		c.WaitForStackCalls = append(c.WaitForStackCalls, call)
		return nil
	} else {
		c.WaitForStackCalls[i].Receives.StackName = stackName
		c.WaitForStackCalls[i].Receives.Pundit = pundit
		return c.WaitForStackCalls[i].Returns.Error
	}
}

func (c *AWSClient) DeleteStack(stackName string) error {
	i := c.DeleteStackCallCount
	c.DeleteStackCallCount++

	if i >= len(c.DeleteStackCalls) {
		call := DeleteStackCall{}
		call.Receives.StackName = stackName
		c.DeleteStackCalls = append(c.DeleteStackCalls, call)
		return nil
	} else {
		c.DeleteStackCalls[i].Receives.StackName = stackName
		return c.DeleteStackCalls[i].Returns.Error
	}
}

func (c *AWSClient) CreateKeyPair(stackName string) (string, error) {
	c.CreateKeyPairCall.Receives.StackName = stackName
	return c.CreateKeyPairCall.Returns.KeyPair, c.CreateKeyPairCall.Returns.Error
}

func (c *AWSClient) DeleteKeyPair(stackName string) error {
	c.DeleteKeyPairCall.Receives.StackName = stackName
	return c.DeleteKeyPairCall.Returns.Error
}

func (c *AWSClient) GetBaseStackResources(stackName string) (awsclient.BaseStackResources, error) {
	c.GetBaseStackResourcesCall.Receives.StackName = stackName
	return c.GetBaseStackResourcesCall.Returns.Resources, c.GetBaseStackResourcesCall.Returns.Error
}

func (c *AWSClient) GetStackResources(stackName string) (map[string]string, error) {
	i := c.GetStackResourcesCallCount
	c.GetStackResourcesCallCount++

	c.GetStackResourcesCalls[i].Receives.StackName = stackName
	return c.GetStackResourcesCalls[i].Returns.Resources, c.GetStackResourcesCalls[i].Returns.Error
}

func (c *AWSClient) CreateAccessKey(userName string) (string, string, error) {
	c.CreateAccessKeyCall.Receives.UserName = userName
	return c.CreateAccessKeyCall.Returns.AccessKey, c.CreateAccessKeyCall.Returns.SecretKey, c.CreateAccessKeyCall.Returns.Error
}

func (c *AWSClient) DeleteAccessKey(userName, accessKeyID string) error {
	c.DeleteAccessKeyCall.Receives.UserName = userName
	c.DeleteAccessKeyCall.Receives.AccessKey = accessKeyID
	return c.DeleteAccessKeyCall.Returns.Error
}

func (c *AWSClient) ListAccessKeys(userName string) ([]string, error) {
	c.ListAccessKeysCall.Receives.UserName = userName
	return c.ListAccessKeysCall.Returns.AccessKeys, c.ListAccessKeysCall.Returns.Error
}
