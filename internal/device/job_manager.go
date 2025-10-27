package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"lumina/internal/dao"
	"lumina/internal/device/exector"
	"lumina/internal/device/metadata"
	"lumina/internal/model"
)

const (
	fetchJobsPath    = "/api/v1/device/jobs"
	reportStatusPath = "/api/v1/device/report-status"
)

func (a *Device) reportDeviceStatus() error {
	a.logger.Debug("report device status")

	jobs, err := a.db.GetJobs()
	if err != nil {
		return err
	}

	deviceStatus := dao.DeviceStatus{
		JobStatus: make(map[string]dao.DeviceJobStatus),
	}

	for _, job := range jobs {
		jobUuid := job.Uuid
		executor, exists := a.executors[jobUuid]
		if !exists {
			deviceStatus.JobStatus[jobUuid] = dao.DeviceJobStatus{
				ExectorStatus: model.ExectorStatusStopped,
			}
		} else {
			deviceStatus.JobStatus[jobUuid] = dao.DeviceJobStatus{
				ExectorStatus: executor.Status(),
			}
		}
	}

	info, err := a.db.GetDeviceInfo()
	if err != nil {
		return err
	} else if info == nil || info.Token == nil {
		return errors.New("device token is nil, please register device")
	}

	url, err := url.Parse(fmt.Sprintf(a.conf.LuminaServerAddr + reportStatusPath))
	if err != nil {
		return err
	}
	body, _ := json.Marshal(deviceStatus)
	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", *info.Token)},
			"Content-Type":  []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(body)),
	}

	resp, err := a.httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http request failed, status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *Device) fetchJobsFromServer(info *metadata.DeviceInfo, lastFetchTs int64) (*dao.ListJobsResponse, error) {
	a.logger.Debugf("fetch jobs, lastFetch: %s", time.Unix(lastFetchTs, 0).Format(time.RFC1123))

	url, err := url.Parse(fmt.Sprintf(a.conf.LuminaServerAddr + fetchJobsPath))
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
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed, status code: %d", resp.StatusCode)
	}

	var respBody dao.ListJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (a *Device) syncJobsFromServer() error {
	info, err := a.db.GetDeviceInfo()
	if err != nil {
		return err
	} else if info == nil || info.Uuid == nil {
		return errors.New("device Id is nil, please register device")
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
			a.logger.Infof("job %s synced", newJob.Uuid)
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

func (a *Device) syncJobsFromMedadata() error {
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
		if !job.Enabled {
			continue
		}
		if _, ok := a.executors[job.Uuid]; !ok {
			a.logger.Infof("job %s created, start the executor", job.Uuid)
			newExector, err := a.newExector(job)
			if err != nil {
				a.logger.WithError(err).Errorf("create job %s executor failed", job.Uuid)
				continue
			}
			if err := newExector.Start(); err != nil {
				a.logger.WithError(err).Errorf("start job %s executor failed", job.Uuid)
			} else {
				a.executors[job.Uuid] = newExector
			}
		}
	}

	return nil
}

func (a *Device) newExector(job *dao.JobSpec) (exector.Executor, error) {
	switch job.Kind {
	case model.JobKindDetect:
		return exector.NewDetector(a.conf, a.deviceInfo, a.ctx, a.minioCli, a.nsqProducer, job)
	case model.JobKindVideoSegment:
		return exector.NewVideoSegmentor(a.conf, a.deviceInfo, a.ctx, a.minioCli, a.nsqProducer, job)
	default:
		return nil, fmt.Errorf("unknown job kind %s", job.Kind)
	}
}
