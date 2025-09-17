package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"lumina/internal/dao"
	"lumina/internal/model"
	"lumina/internal/utils"
)

const fetchJobsPath = "/api/v1/agent/%s/jobs"

func (a *Agent) fetchJobsFromServer(info *AgentInfo, lastFetchTs int64) (*dao.GetJobListResp, error) {
	a.logger.Debugf("fetch jobs, lastFetch: %s", time.Unix(lastFetchTs, 0).Format(time.RFC1123))

	url, err := url.Parse(fmt.Sprintf(a.conf.LuminaServerAddr+fetchJobsPath, *info.Uuid))
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
	if lastFetchTs > 0 {
		req.Header.Set("If-Modified-Since", time.Unix(lastFetchTs, 0).Format(time.RFC1123))
	}

	resp, err := a.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		a.logger.Debug("no new jobs")
		return nil, nil
	}

	var respBody dao.GetJobListResp
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (a *Agent) syncJobsFromServer() error {
	info, err := a.db.GetAgentInfo()
	if err != nil {
		return err
	} else if info == nil || info.Uuid == nil {
		return errors.New("agent Id is nil, please register agent")
	}

	lastFetchTs, err := a.db.GetLastFetchTime()
	if err != nil {
		return err
	}

	resp, err := a.fetchJobsFromServer(info, lastFetchTs)
	if err != nil {
		return err
	} else if resp == nil {
		a.logger.Debug("no need to update jobs")
		return nil
	}

	newJobs := make(map[string]*dao.JobSpec)
	maxTime := time.Unix(0, 0)
	for _, job := range resp.Items {
		newJobs[job.Uuid] = &job

		updateTime, err2 := time.Parse(time.RFC3339, job.UpdateTime)
		if err2 != nil {
			a.logger.WithError(err2).Error("parse job update time")
		}
		if updateTime.After(maxTime) {
			maxTime = updateTime
		}
	}

	jobs, err := a.db.GetJobs()
	if err != nil {
		return err
	}
	oldJobs := make(map[string]*dao.JobSpec)
	for _, job := range jobs {
		oldJobs[job.Uuid] = job
	}

	allDbSynced := true
	for _, oldJob := range oldJobs {
		if newJob, ok := newJobs[oldJob.Uuid]; !ok {
			a.logger.Infof("job %s deleted", oldJob.Uuid)
			if err := a.db.DeleteJob(oldJob.Uuid); err != nil {
				a.logger.WithError(err).Errorf("delete job %s failed", oldJob.Uuid)
				allDbSynced = false
			}
		} else if newJob.UpdateTime != oldJob.UpdateTime {
			a.logger.Infof("job %s updated", oldJob.Uuid)
			if err := a.db.SetJob(oldJob.Uuid, newJob); err != nil {
				a.logger.WithError(err).Errorf("update job %s failed", oldJob.Uuid)
				allDbSynced = false
			}
		}
	}
	for _, newJob := range newJobs {
		if _, ok := oldJobs[newJob.Uuid]; !ok {
			a.logger.Infof("job %s created", newJob.Uuid)
			if err := a.db.SetJob(newJob.Uuid, newJob); err != nil {
				a.logger.WithError(err).Errorf("create job %s failed", newJob.Uuid)
				allDbSynced = false
			}
		}
	}

	// if db sync failed, do not update last fetch time, try next time
	if allDbSynced {
		a.db.SetLastFetchTime(maxTime.Unix())
	}

	return nil
}

func (a *Agent) syncJobsFromMedadata(reclaimCh chan string) error {
	jobs, err := a.db.GetJobs()
	if err != nil {
		return err
	}
	metaJobs := make(map[string]*dao.JobSpec)
	for _, job := range jobs {
		metaJobs[job.Uuid] = job
	}

	for _, e := range a.executors {
		job := e.Job()
		if metaJob, ok := metaJobs[job.Uuid]; !ok {
			a.logger.Infof("job %s deleted, stop the executor", job.Uuid)
			e.Stop()
			delete(a.executors, job.Uuid)
		} else if metaJob.UpdateTime != job.UpdateTime {
			a.logger.Infof("job %s updated, stop the executor", job.Uuid)
			e.Stop()
			delete(a.executors, job.Uuid)
		}
	}

	for _, job := range metaJobs {
		if job.Status == model.JobStatusStopped {
			continue
		}
		if _, ok := a.executors[job.Uuid]; !ok {
			a.logger.Infof("job %s created, start the executor", job.Uuid)
			newExector, err := NewDetector(a.tritonCli, a.conf.JobDir(), a.ctx, job)
			if err != nil {
				a.logger.WithError(err).Errorf("create job %s executor failed", job.Uuid)
				continue
			}
			if err := newExector.Start(); err != nil {
				a.logger.WithError(err).Errorf("start job %s executor failed", job.Uuid)
			} else {
				a.executors[job.Uuid] = newExector
				go func() {
					<-newExector.Done()
					reclaimCh <- job.Uuid
				}()
			}
		}
	}

	return nil
}

func (a *Agent) uploadRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var minioCli *minio.Client
	var info *AgentInfo
	lastInfoUpdateTime := time.Now()

	for {
		if minioCli == nil || time.Since(lastInfoUpdateTime) > time.Hour {
			var err error
			info, err = a.db.GetAgentInfo()
			if err != nil {
				a.logger.WithError(err).Errorf("get agent info failed")
				continue
			} else if info == nil {
				continue
			}

			region := a.conf.S3.Region
			if region == "" {
				region = "us-east-1"
			}
			minioCli, err = minio.New(a.conf.S3.Endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(*info.S3AccessKeyID, *info.S3SecretAccessKey, ""),
				Secure: a.conf.S3.UseSSL,
				Region: region,
			})
			if err != nil {
				a.logger.WithError(err).Errorf("create minio client failed")
				continue
			}
			lastInfoUpdateTime = time.Now()
		}

		if err := a.listAndUpload(minioCli, info); err != nil {
			a.logger.WithError(err).Errorf("list and upload failed")
		}

		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *Agent) listAndUpload(minioCli *minio.Client, info *AgentInfo) error {
	jobDir := a.conf.JobDir()
	entries, err := os.ReadDir(jobDir)
	if err != nil {
		if os.IsNotExist(err) {
			a.logger.Debug("job directory does not exist")
			return nil
		}
		return fmt.Errorf("read job directory failed: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		jobID := entry.Name()
		jobPath := filepath.Join(jobDir, jobID)

		err := filepath.WalkDir(jobPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
				return nil
			}

			jsonData, err := os.ReadFile(path)
			if err != nil {
				a.logger.WithError(err).Errorf("read JSON file %s failed", path)
				return nil
			}

			var result dao.DetectionResult
			if err := json.Unmarshal(jsonData, &result); err != nil {
				a.logger.WithError(err).Errorf("unmarshal JSON file %s failed", path)
				return nil
			}

			fileName := strings.TrimSuffix(d.Name(), ".json")
			imgPath := filepath.Join(jobPath, fileName+".jpg")

			var ts time.Time
			if result.Timestamp != 0 {
				ts = time.Unix(result.Timestamp/1000000000, result.Timestamp%1000000000)
			} else {
				if jsonInfo, err := d.Info(); err == nil {
					ts = jsonInfo.ModTime()
				} else {
					ts = time.Now()
				}
			}
			minioPath := fmt.Sprintf("/%s/%04d/%02d/%02d/%s/%s.jpg",
				*info.Uuid, ts.Year(), ts.Month(), ts.Day(), result.JobId, fileName)

			ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
			defer cancel()
			if err := utils.UploadImageToMinio(ctx, minioCli, a.conf.S3.Bucket, imgPath, minioPath); err != nil {
				a.logger.WithError(err).Errorf("upload image %s to minio failed", imgPath)
				return nil
			}

			msg := &dao.Message{
				JobUuid:     result.JobId,
				Timestamp:   ts.UnixNano(),
				ImagePath:   minioPath,
				DetectBoxes: result.Boxes,
			}
			msgData, _ := json.Marshal(msg)
			if err := a.nsqProducer.Publish(a.conf.NSQ.Topic, msgData); err != nil {
				a.logger.WithError(err).Errorf("publish to NSQ failed for %s", path)
				return nil
			}

			os.Remove(path)
			os.Remove(imgPath)

			a.logger.Infof("successfully processed %s: uploaded image to %s and sent to NSQ", path, minioPath)
			return nil
		})
		if err != nil {
			a.logger.WithError(err).Errorf("process job directory %s failed", jobPath)
			continue
		}
	}

	return nil
}
