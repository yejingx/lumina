package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
)

// handleJobStats 任务统计
// @Summary 获取任务统计
// @Description 根据job_id从InfluxDB查询消息数量趋势；检测任务还返回各Label数量趋势
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Param start query string false "开始时间(RFC3339)"
// @Param end query string false "结束时间(RFC3339)"
// @Param window query string false "聚合窗口，如1m、5m、15m" default(5m)
// @Success 200 {object} dao.JobStatsResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id}/stats [get]
func (s *Server) handleJobStats(c *gin.Context) {
	if s.influxQuery == nil || !s.conf.InfluxDB.Enabled {
		s.writeError(c, http.StatusBadRequest, fmt.Errorf("influxdb not enabled"))
		return
	}

	job := c.MustGet(jobKey).(*model.Job)

	var req dao.JobStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	end := time.Now().UTC()
	if req.End != "" {
		te, err := time.Parse(time.RFC3339, req.End)
		if err != nil {
			s.writeError(c, http.StatusBadRequest, fmt.Errorf("invalid end: %w", err))
			return
		}
		end = te.UTC()
	}

	start := end.Add(-24 * time.Hour)
	if req.Start != "" {
		ts, err := time.Parse(time.RFC3339, req.Start)
		if err != nil {
			s.writeError(c, http.StatusBadRequest, fmt.Errorf("invalid start: %w", err))
			return
		}
		start = ts.UTC()
	}
	if !start.Before(end) {
		s.writeError(c, http.StatusBadRequest, fmt.Errorf("start must be before end"))
		return
	}

	window := req.Window
	if window == "" {
		window = "5m"
	}
	if !isValidWindow(window) {
		s.writeError(c, http.StatusBadRequest, fmt.Errorf("invalid window: %s", window))
		return
	}

	messages, err := s.queryMessagesTrend(c.Request.Context(), job.Uuid, start, end, window)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.JobStatsResponse{
		Messages: messages,
	}

	if job.Kind == model.JobKindDetect {
		labels, err := s.queryLabelsTrend(c.Request.Context(), job.Uuid, start, end, window)
		if err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		}
		resp.Labels = labels
	}

	c.JSON(http.StatusOK, resp)
}

func isValidWindow(w string) bool {
	re := regexp.MustCompile(`^[0-9]+(ms|s|m|h|d|w)$`)
	return re.MatchString(w)
}

const influxMeasurementMessage = "lumina_message"
const influxMeasurementDetection = "lumina_detection"

func (s *Server) queryMessagesTrend(ctx context.Context, jobUuid string, start, end time.Time, window string) ([]dao.TimeCount, error) {
	flux := fmt.Sprintf(
		`from(bucket: "%s")
      |> range(start: time(v: "%s"), stop: time(v: "%s"))
      |> filter(fn: (r) => r["_measurement"] == "%s")
      |> filter(fn: (r) => r["job_uuid"] == "%s")
      |> filter(fn: (r) => r["_field"] == "count")
      |> aggregateWindow(every: %s, fn: count, createEmpty: false)`,
		s.conf.InfluxDB.Bucket,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		influxMeasurementMessage,
		jobUuid,
		window,
	)

	res, err := s.influxQuery.Query(ctx, flux)
	if err != nil {
		return nil, fmt.Errorf("query messages trend: %w", err)
	}
	defer res.Close()

	items := make([]dao.TimeCount, 0, 32)
	for res.Next() {
		rec := res.Record()
		t := rec.Time().UTC().Format(time.RFC3339)
		count := toInt64(rec.Value())
		items = append(items, dao.TimeCount{Time: t, Count: count})
	}
	if res.Err() != nil {
		return nil, fmt.Errorf("query messages trend result error: %v", res.Err())
	}
	return items, nil
}

func (s *Server) queryLabelsTrend(ctx context.Context, jobUuid string, start, end time.Time, window string) ([]dao.LabelTimeCount, error) {
	flux := fmt.Sprintf(
		`from(bucket: "%s")
      |> range(start: time(v: "%s"), stop: time(v: "%s"))
      |> filter(fn: (r) => r["_measurement"] == "%s")
      |> filter(fn: (r) => r["job_uuid"] == "%s")
      |> filter(fn: (r) => r["_field"] == "confidence")
      |> aggregateWindow(every: %s, fn: count, createEmpty: false)
      |> group(columns: ["label"])`,
		s.conf.InfluxDB.Bucket,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		influxMeasurementDetection,
		jobUuid,
		window,
	)

	res, err := s.influxQuery.Query(ctx, flux)
	if err != nil {
		return nil, fmt.Errorf("query labels trend: %w", err)
	}
	defer res.Close()

	items := make([]dao.LabelTimeCount, 0, 64)
	for res.Next() {
		rec := res.Record()
		label, _ := rec.ValueByKey("label").(string)
		t := rec.Time().UTC().Format(time.RFC3339)
		count := toInt64(rec.Value())
		items = append(items, dao.LabelTimeCount{Label: label, Time: t, Count: count})
	}
	if res.Err() != nil {
		return nil, fmt.Errorf("query labels trend result error: %v", res.Err())
	}
	return items, nil
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case uint64:
		return int64(t)
	case int32:
		return int64(t)
	case uint32:
		return int64(t)
	case float64:
		return int64(t)
	case float32:
		return int64(t)
	case int:
		return int64(t)
	case uint:
		return int64(t)
	default:
		return 0
	}
}
