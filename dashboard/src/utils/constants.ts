// 分页默认配置
export const DEFAULT_PAGE_SIZE = 20;
export const DEFAULT_PAGE_SIZE_OPTIONS = ['10', '20', '50', '100'];

// 状态映射
export const JOB_STATUS_MAP = {
  stopped: { text: '已停止', color: 'default' },
  running: { text: '运行中', color: 'processing' },
};

// 任务类型映射
export const JOB_KIND_MAP = {
  detect: { text: '检测任务', color: 'blue' },
  video_segment: { text: '视频分割', color: 'green' },
};

export const DEVICE_STATUS_MAP = {
  online: { text: '在线', color: 'success' },
  offline: { text: '离线', color: 'default' },
  error: { text: '错误', color: 'error' },
};

// 消息类型映射
export const MESSAGE_TYPE_MAP = {
  info: { text: '信息', color: 'blue' },
  warning: { text: '警告', color: 'orange' },
  error: { text: '错误', color: 'red' },
  success: { text: '成功', color: 'green' },
};

// 日期格式
export const DATE_FORMAT = 'YYYY-MM-DD HH:mm:ss';
export const DATE_FORMAT_SHORT = 'YYYY-MM-DD';