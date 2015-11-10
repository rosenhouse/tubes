package mocks

import "time"

type SleepCall struct {
	Receives struct {
		Duration time.Duration
	}
}

type Clock struct {
	SleepCalls []SleepCall
}

func (c *Clock) Sleep(duration time.Duration) {
	call := SleepCall{}
	call.Receives.Duration = duration
	c.SleepCalls = append(c.SleepCalls, call)
}
