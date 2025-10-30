package device

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"lumina/internal/dao"
	"lumina/internal/device/metadata"
)

const fetchPreviewTasksPath = "/api/v1/device/preview-tasks"

type PreviewJob struct {
	Task   dao.PreviewTask `json:"task"`
	ctx    context.Context
	cancel context.CancelFunc
}

func (j *PreviewJob) Cancel() {
	if j.cancel != nil {
		j.cancel()
	}
}

func (a *Device) fetchPreviewTasksFromServer(info *metadata.DeviceInfo) (*dao.ListPreviewTasksResponse, error) {
	a.logger.Debugf("fetch preview tasks")

	url, err := url.Parse(fmt.Sprintf(a.conf.LuminaServerAddr + fetchPreviewTasksPath))
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", *info.Token)},
		},
	}

	resp, err := a.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed, status code: %d", resp.StatusCode)
	}

	var respBody dao.ListPreviewTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (a *Device) syncPreviewTasksFromServer() error {
	for _, job := range a.previewJobs {
		if job.Task.Expired() {
			a.logger.Infof("preview task expired, task uuid: %s", job.Task.TaskUuid)
			job.Cancel()
			delete(a.previewJobs, job.Task.TaskUuid)
		}
	}

	info, err := a.db.GetDeviceInfo()
	if err != nil {
		return err
	} else if info == nil || info.Uuid == nil {
		return errors.New("device Id is nil, please register device")
	}

	resp, err := a.fetchPreviewTasksFromServer(info)
	if err != nil {
		return err
	} else if resp == nil {
		a.logger.Debug("no need to update jobs")
		return nil
	}

	newPreviewTasks := make(map[string]dao.PreviewTask)
	for _, task := range resp.Items {
		newPreviewTasks[task.TaskUuid] = task
		_, exist := a.previewJobs[task.TaskUuid]
		if !exist {
			a.logger.Infof("start new preview task, task: %+v", task)
			a.previewJobs[task.TaskUuid] = a.startPreviewJob(a.ctx, &task)
		}
	}

	for _, job := range a.previewJobs {
		if _, ok := newPreviewTasks[job.Task.TaskUuid]; !ok {
			a.logger.Infof("stop preview task, task uuid: %s", job.Task.TaskUuid)
			job.Cancel()
			delete(a.previewJobs, job.Task.TaskUuid)
		}
	}

	return nil
}

func (a *Device) startPreviewJob(ctx context.Context, task *dao.PreviewTask) *PreviewJob {
	job := &PreviewJob{
		Task: *task,
	}
	job.ctx, job.cancel = context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-job.ctx.Done():
				a.logger.WithField("taskUuid", task.TaskUuid).Debugf("preview job canceled")
				return
			default:
			}
			// 先探测输入流的编码格式
			codec, err := probeVideoCodec(job.ctx, task.PullAddr)
			if err != nil {
				a.logger.WithField("taskUuid", task.TaskUuid).Warnf("ffprobe failed, default to copy, err: %v", err)
			} else {
				a.logger.WithField("taskUuid", task.TaskUuid).Infof("detected input codec: %s", codec)
			}

			// 根据编码选择是否转码
			args := []string{"-i", task.PullAddr, "-an"}
			switch strings.ToLower(codec) {
			case "h264":
				args = append(args, "-c:v", "copy")
			case "hevc", "h265":
				args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency", "-pix_fmt", "yuv420p")
			default:
				// 未知编码时默认尝试直接复制
				args = append(args, "-c:v", "copy")
			}
			args = append(args, "-f", "flv", task.PushAddr)

			cmd := exec.CommandContext(job.ctx, "ffmpeg", args...)
			err = cmd.Run()
			if err != nil {
				a.logger.WithField("taskUuid", task.TaskUuid).Errorf("preview job failed, err: %v", err)
				time.Sleep(5 * time.Second)
			}
			a.logger.Debugf("preview job exit, task uuid: %s", task.TaskUuid)
		}
	}()

	return job
}

// 使用 ffprobe 探测视频编码
func probeVideoCodec(ctx context.Context, input string) (string, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	codec := strings.TrimSpace(out.String())
	return codec, nil
}
