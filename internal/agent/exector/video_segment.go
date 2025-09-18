package exector

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"lumina/internal/dao"
	"lumina/pkg/log"
)

type VideoSegmentor struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	job     *dao.JobSpec
	logger  *logrus.Entry
	status  ExectorStatus
	workDir string
}

func NewVideoSegmentor(workDir string, parentCtx context.Context, job *dao.JobSpec) (*VideoSegmentor, error) {
	if job.VideoSegment == nil {
		return nil, fmt.Errorf("job %s video segment is nil", job.Uuid)
	}

	workDir = path.Join(workDir, job.Uuid)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &VideoSegmentor{
		ctx:     ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
		job:     job,
		logger:  log.GetLogger(ctx).WithField("job", job.Uuid),
		status:  ExectorStatusStopped,
		workDir: workDir,
	}, nil
}

func (e *VideoSegmentor) Job() *dao.JobSpec {
	return e.job
}

func (e *VideoSegmentor) Status() ExectorStatus {
	return e.status
}

func (e *VideoSegmentor) Start() error {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.logger.Info("video segmentation job started")
		e.status = ExectorStatusRunning
		e.runJob()
		e.logger.Info("video segmentation job finished")
	}()
	return nil
}

func (e *VideoSegmentor) Stop() {
	e.cancel()
	e.wg.Wait()
	e.status = ExectorStatusStopped
}

func (e *VideoSegmentor) runJob() {
	interval := 30
	if e.job.VideoSegment.Interval > 0 {
		interval = e.job.VideoSegment.Interval
	}

	startTs := time.Now().Format("20060102150405")
	outputPattern := "segment_" + startTs + "_%06d.mp4"

	args := []string{
		"-i", e.job.Input, // 输入视频流
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
		e.status = ExectorStatusFailed
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
		e.status = ExectorStatusStopped
	case err := <-done:
		if err != nil {
			e.logger.WithError(err).Error("ffmpeg process exited with error")
			e.status = ExectorStatusFailed
		} else {
			e.logger.Info("ffmpeg process completed successfully")
			e.status = ExectorStatusFinished
		}
	}
}
