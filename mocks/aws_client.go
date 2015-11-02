package mocks

type AWSClient struct {
	GetLatestNATBoxAMIIDCall struct {
		Returns struct {
			AMIID string
			Error error
		}
	}
}

func (c *AWSClient) GetLatestNATBoxAMIID() (string, error) {
	return c.GetLatestNATBoxAMIIDCall.Returns.AMIID, c.GetLatestNATBoxAMIIDCall.Returns.Error
}
