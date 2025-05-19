package types

import "context"

type StatusCode uint

const (
	StatusScheduled StatusCode = 1
	StatusRunning   StatusCode = 2
	StatusComplete  StatusCode = 3
	StatusError     StatusCode = 4
)

type Scheduler interface {
	Schedule(Job)
	Start()
	Shutdown()
}

type Job interface {
	Run(ctx context.Context) error
	Status() StatusCode
}
