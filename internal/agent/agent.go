package agent

import (
	"context"
	"lumina/pkg/log"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	conf   *Config
	ctx    context.Context
	logger *logrus.Entry
	db     *badger.DB
}

func NewAgent(conf *Config) (*Agent, error) {
	ctx := context.Background()
	logger := log.GetLogger(ctx).WithField("component", "agent")

	db, err := badger.Open(badger.DefaultOptions(conf.DataDir()))
	if err != nil {
		return nil, err
	}

	return &Agent{
		conf:   conf,
		ctx:    ctx,
		logger: logger,
		db:     db,
	}, nil
}

func (a *Agent) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.logger.Info("tick")
		}
	}
}

func (a *Agent) Stop() {
	<-a.ctx.Done()
	a.db.Close()
	a.logger.Info("agent stopped")
}
