package mocks

type PunditCall struct {
	Receives struct {
		StatusString string
	}
	Returns struct {
		Result bool
	}
}

type CloudFormationStatusPundit struct {
	IsHealthyCalls      []PunditCall
	IsHealthyCallCount  int
	IsCompleteCalls     []PunditCall
	IsCompleteCallCount int
}

func NewCloudFormationStatusPundit(nCalls int) *CloudFormationStatusPundit {
	return &CloudFormationStatusPundit{
		IsHealthyCalls:  make([]PunditCall, nCalls),
		IsCompleteCalls: make([]PunditCall, nCalls),
	}
}

func (p *CloudFormationStatusPundit) IsHealthy(statusString string) bool {
	i := p.IsHealthyCallCount
	p.IsHealthyCalls[i].Receives.StatusString = statusString
	result := p.IsHealthyCalls[i].Returns.Result
	p.IsHealthyCallCount++
	return result
}

func (p *CloudFormationStatusPundit) IsComplete(statusString string) bool {
	i := p.IsCompleteCallCount
	p.IsCompleteCalls[i].Receives.StatusString = statusString
	result := p.IsCompleteCalls[i].Returns.Result
	p.IsCompleteCallCount++
	return result
}
