package dao

// JobStatsRequest 查询参数
// 采用 RFC3339 时间字符串和窗口字符串（如 1m、5m、15m）
// 若未提供则使用默认值：start=过去24小时, end=当前时间, window=5m
type JobStatsRequest struct {
	Start  string `form:"start" json:"start"`
	End    string `form:"end" json:"end"`
	Window string `form:"window" json:"window"`
}

// TimeCount 用于消息数量趋势
type TimeCount struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

// LabelTimeCount 用于各分类标签数量趋势
type LabelTimeCount struct {
	Label string `json:"label"`
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

// JobStatsResponse 服务端返回结构
// detect 任务返回 messages + labels；video_segment 仅返回 messages
type JobStatsResponse struct {
	Messages []TimeCount      `json:"messages"`
	Labels   []LabelTimeCount `json:"labels,omitempty"`
}
