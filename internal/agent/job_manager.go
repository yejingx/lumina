package agent

import "lumina/internal/dao"

const fetchJobsPath = "/api/v1/agent/%s/jobs"

func (a *Agent) fetchJobs() ([]dao.JobSpec, error) {
	a.logger.Debug("fetch jobs")
	return nil, nil
}
