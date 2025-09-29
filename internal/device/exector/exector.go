package exector

import (
	"lumina/internal/dao"
)

type ExectorStatus int

const (
	ExectorStatusStopped ExectorStatus = iota
	ExectorStatusRunning
	ExectorStatusFinished
	ExectorStatusFailed
)

type Executor interface {
	Start() error
	Stop()
	Job() *dao.JobSpec
	Status() ExectorStatus
}
