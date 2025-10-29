package exector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"

	"lumina/internal/dao"
	"lumina/internal/device/config"
	"lumina/internal/device/metadata"
	"lumina/internal/model"
	"lumina/internal/utils"
	"lumina/pkg/log"
)

type VideoSegmentor struct {
	ctx         context.Context
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
	job         *dao.JobSpec
	logger      *logrus.Entry
	status      model.ExectorStatus
	workDir     string
	conf        *config.Config
	nsqProducer *nsq.Producer
	minioCli    *minio.Client
	deviceInfo  *metadata.DeviceInfo
}

func NewVideoSegmentor(conf *config.Config, deviceInfo *metadata.DeviceInfo, parentCtx context.Context,
	minioCli *minio.Client, nsqProducer *nsq.Producer, job *dao.JobSpec) (*VideoSegmentor, error) {
	if job.VideoSegment == nil {
		return nil, fmt.Errorf("job %s video segment is nil", job.Uuid)
	}
	workDir := path.Join(conf.JobDir(), job.Uuid)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &VideoSegmentor{
		deviceInfo:  deviceInfo,
		ctx:         ctx,
		cancel:      cancel,
		wg:          &sync.WaitGroup{},
		job:         job,
		logger:      log.GetLogger(ctx).WithField("job", job.Uuid),
		status:      model.ExectorStatusStopped,
		workDir:     workDir,
		conf:        conf,
		nsqProducer: nsqProducer,
		minioCli:    minioCli,
	}, nil
}

func (e *VideoSegmentor) Job() *dao.JobSpec {
	return e.job
}

func (e *VideoSegmentor) Status() model.ExectorStatus {
	return e.status
}

func (e *VideoSegmentor) Start() error {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.logger.Info("video segmentation job started")
		e.status = model.ExectorStatusRunning
		e.runJob()
		e.logger.Info("video segmentation job finished")
	}()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.uploadRoutine()
	}()

	return nil
}

func (e *VideoSegmentor) Stop() {
	e.cancel()
	e.wg.Wait()
	e.status = model.ExectorStatusStopped
}

func (e *VideoSegmentor) runJob() {
	interval := 30
	if e.job.VideoSegment.Interval > 0 {
		interval = e.job.VideoSegment.Interval
	}

	outputPattern := "segment_%Y%m%d_%H%M%S.mp4"

	args := []string{
		"-i", e.job.Input(), // 输入视频流
		"-c", "copy", // 复制编码，不重新编码
		"-f", "segment", // 使用 segment 格式
		"-segment_time", fmt.Sprintf("%d", interval), // 分段时间间隔
		"-segment_format", "mp4", // 分段格式
		"-reset_timestamps", "1", // 重置时间戳
		"-strftime", "1", // 启用时间格式化
		outputPattern, // 输出文件模式
	}

	e.logger.WithFields(logrus.Fields{
		"args": args,
	}).Info("starting ffmpeg video segmentation")

	cmd := exec.CommandContext(e.ctx, "ffmpeg", args...)
	cmd.Dir = e.workDir

	if err := cmd.Start(); err != nil {
		e.logger.WithError(err).Error("failed to start ffmpeg process")
		e.status = model.ExectorStatusFailed
		return
	}

	e.logger.WithField("pid", cmd.Process.Pid).Info("ffmpeg process started")

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-e.ctx.Done():
		e.logger.Info("context cancelled, terminating ffmpeg process")
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				e.logger.WithError(err).Error("failed to kill ffmpeg process")
			} else {
				e.logger.Info("ffmpeg process terminated")
			}
		}
		e.status = model.ExectorStatusStopped
	case err := <-done:
		if err != nil {
			e.logger.WithError(err).Error("ffmpeg process exited with error")
			e.status = model.ExectorStatusFailed
		} else {
			e.logger.Info("ffmpeg process completed successfully")
			e.status = model.ExectorStatusFinished
		}
	}
}

func (e *VideoSegmentor) uploadRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		if err := e.listAndUpload(); err != nil {
			e.logger.WithError(err).Errorf("list and upload failed")
		}

		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (e *VideoSegmentor) listAndUpload() error {
	var files []string
	err := filepath.WalkDir(e.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".mp4") {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(files)

	if len(files) <= 1 {
		return nil
	}

	// 处理除最后一个文件外的所有文件
	for _, path := range files[:len(files)-1] {
		filename := filepath.Base(path)
		info, err := os.Stat(path)
		if err != nil {
			e.logger.WithError(err).Warnf("failed to get file info %s, skip", filename)
			continue
		}
		ts := info.ModTime()
		minioPath := fmt.Sprintf("/%s/%04d/%02d/%02d/%s/%s",
			*e.deviceInfo.Uuid, ts.Year(), ts.Month(), ts.Day(), e.job.Uuid, filename)

		// 上传到 MinIO
		ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
		if err := utils.UploadFileToMinio(ctx, e.minioCli, e.conf.S3.Bucket, path, minioPath); err != nil {
			e.logger.WithError(err).Errorf("upload video segment %s to minio failed", path)
			cancel()
			continue
		}
		cancel()

		// 创建消息并发送到 NSQ
		msg := &dao.DeviceMessage{
			JobUuid:   e.job.Uuid,
			Timestamp: ts.UnixNano(),
			VideoPath: minioPath,
		}
		msgData, _ := json.Marshal(msg)
		if err := e.nsqProducer.Publish(e.conf.NSQ.Topic, msgData); err != nil {
			e.logger.WithError(err).Errorf("publish to NSQ failed for %s", path)
			continue
		}

		// 删除本地文件
		if err := os.Remove(path); err != nil {
			e.logger.WithError(err).Warnf("failed to remove local file %s", path)
		}

		e.logger.Infof("successfully processed %s: uploaded to %s and sent to NSQ", path, minioPath)
	}

	return nil
}
