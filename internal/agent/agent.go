package agent

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/Trendyol/go-triton-client/base"
	tritonGrpc "github.com/Trendyol/go-triton-client/client/grpc"
	"github.com/sirupsen/logrus"

	"lumina/pkg/log"
)

type Agent struct {
	conf      *Config
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *logrus.Entry
	db        *MetadataDB
	httpCli   *http.Client
	tritonCli base.Client
	executors map[string]*JobExecutor
}

func NewAgent(conf *Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := log.GetLogger(ctx).WithField("component", "agent")

	db, err := NewMetadataDB(conf.DataDir(), logger)
	if err != nil {
		cancel()
		return nil, err
	}

	httpCli := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 15 * time.Second,
	}

	tritonCli, err := tritonGrpc.NewClient(
		conf.Triton.ServerAddr,
		false, // verbose logging
		30,    // connection timeout in seconds
		30,    // network timeout in seconds
		false, // use SSL
		true,  // insecure connection
		nil,   // existing gRPC connection
		nil,   // logger
	)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Agent{
		conf:      conf,
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		db:        db,
		httpCli:   httpCli,
		tritonCli: tritonCli,
		executors: make(map[string]*JobExecutor),
	}, nil
}

func (a *Agent) Start() {
	fetchTicker := time.NewTicker(5 * time.Second)
	syncTicker := time.NewTicker(1 * time.Second)
	defer func() {
		fetchTicker.Stop()
		syncTicker.Stop()
		a.logger.Info("agent stopped")
	}()

	reclaimCh := make(chan string, 10)

	for {
		select {
		case <-a.ctx.Done():
			return
		case jobId := <-reclaimCh:
			a.logger.Infof("reclaim job %s", jobId)
			delete(a.executors, jobId)
		case <-fetchTicker.C:
			a.logger.Debug("fetch tick")
			if err := a.syncJobsFromServer(); err != nil {
				a.logger.WithError(err).Errorf("sync jobs from server failed")
			}
		case <-syncTicker.C:
			a.logger.Debug("sync tick")
			if err := a.syncJobsFromMedadata(reclaimCh); err != nil {
				a.logger.WithError(err).Errorf("sync jobs from metadata failed")
			}
		}
	}
}

func (a *Agent) Stop() {
	a.cancel()
	a.db.Close()
}
