package exector

import (
	"lumina/internal/dao"
	"lumina/internal/model"
)

type Executor interface {
	Start() error
	Stop()
	Job() *dao.JobSpec
	Status() model.ExectorStatus
}
