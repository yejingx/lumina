package device

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"

	"lumina/internal/device/config"
	"lumina/internal/device/exector"
	"lumina/internal/device/metadata"
	"lumina/pkg/log"
)

type Device struct {
	conf        *config.Config
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logrus.Entry
	db          *metadata.MetadataDB
	httpCli     *http.Client
	executors   map[string]exector.Executor
	deviceInfo  *metadata.DeviceInfo
	nsqProducer *nsq.Producer
	minioCli    *minio.Client
}

func NewDevice(conf *config.Config) (*Device, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := log.GetLogger(ctx).WithField("component", "device")

	db, err := metadata.NewMetadataDB(conf.DataDir(), logger)
	if err != nil {
		cancel()
		return nil, err
	}

	info, err := db.GetDeviceInfo()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("get device info failed: %w", err)
	} else if info == nil {
		cancel()
		return nil, fmt.Errorf("device info is nil")
	}

	region := conf.S3.Region
	if region == "" {
		region = "us-east-1"
	}
	minioCli, err := minio.New(conf.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*info.S3AccessKeyID, *info.S3SecretAccessKey, ""),
		Secure: conf.S3.UseSSL,
		Region: region,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create minio client failed: %w", err)
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

	producer, err := nsq.NewProducer(conf.NSQ.NSQDAddr, nsq.NewConfig())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create NSQ producer failed: %w", err)
	}

	return &Device{
		conf:        conf,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
		db:          db,
		httpCli:     httpCli,
		executors:   make(map[string]exector.Executor),
		deviceInfo:  info,
		nsqProducer: producer,
		minioCli:    minioCli,
	}, nil
}

func (a *Device) Start() {
	fetchTicker := time.NewTicker(5 * time.Second)
	syncTicker := time.NewTicker(1 * time.Second)
	defer func() {
		fetchTicker.Stop()
		syncTicker.Stop()
		a.logger.Info("device stopped")
	}()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-fetchTicker.C:
			a.logger.Debug("fetch tick")
			if err := a.syncJobsFromServer(); err != nil {
				a.logger.WithError(err).Errorf("sync jobs from server failed")
			}
			if err := a.reportDeviceStatus(); err != nil {
				a.logger.WithError(err).Errorf("report device status failed")
			}
		case <-syncTicker.C:
			a.logger.Debug("sync tick")
			if err := a.syncJobsFromMedadata(); err != nil {
				a.logger.WithError(err).Errorf("sync jobs from metadata failed")
			}
		}
	}
}

func (a *Device) Stop() {
	a.cancel()
	a.db.Close()
	a.nsqProducer.Stop()
}
