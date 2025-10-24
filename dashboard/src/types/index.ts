// 基础响应类型
export interface BaseResponse<T = any> {
  data?: T;
  message?: string;
  code?: number;
}

// 分页响应类型
export interface PaginationResponse<T> {
  items: T[];
  total: number;
  start: number;
  limit: number;
}

// 错误响应类型
export interface ErrorResponse {
  error: string;
  message?: string;
}

// 用户类型
export interface User {
  id: number;
  username: string;
  email?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateUserRequest {
  username: string;
  email?: string;
  password: string;
}

export interface CreateUserResponse {
  id: number;
  username: string;
  email?: string;
}

export interface ListUsersResponse {
  users: User[];
  total: number;
}

// 设备类型
export interface Device {
  id: number;
  token: string;
  uuid: string;
  registerTime: string;
  lastPingTime: string;
}

export interface DeviceSpec {
  id: number;
  token: string;
  uuid: string;
  registerTime: string;
  lastPingTime: string;
}

export interface RegisterRequest {
  name: string;
  ip?: string;
  port?: number;
  metadata?: Record<string, any>;
}

export interface RegisterResponse {
  id: number;
  token: string;
}

export interface ListDeviceResponse {
  devices: Device[];
  total: number;
}

// 接入凭证类型
export interface AccessToken {
  id: number;
  accessToken: string;
  createTime: string;
  expireTime: string;
  deviceUuid: string;
}

export interface AccessTokenSpec {
  id: number;
  accessToken: string;
  createTime: string;
  expireTime: string;
  deviceUuid: string;
}

export interface CreateAccessTokenRequest {
  expireTime: string;
}

export interface CreateAccessTokenResponse {
  id: number;
  accessToken: string;
  createTime: string;
  expireTime: string;
  deviceUuid: string;
}

export interface ListAccessTokenResponse {
  accessTokens: AccessToken[];
  total: number;
}

// Job related enums and types
export type JobKind = 'detect' | 'video_segment';
export type JobStatus = 'stopped' | 'running';

export interface DetectOptions {
  modelName: string;
  labels?: string;
  confThreshold?: number;
  iouThreshold?: number;
  interval?: number;
  triggerCount?: number;
  triggerInterval?: number;
}

export interface VideoSegmentOptions {
  interval?: number;
}

// 结果过滤类型
export type Operator =
  | 'eq'
  | 'ne'
  | 'in'
  | 'not_in'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'ends_with'
  | 'empty'
  | 'not_empty';

export type CombineOperator = 'and' | 'or';

export interface Condition {
  field: string;
  op: Operator;
  value?: string;
}

export interface FilterCondition {
  combineOp: CombineOperator;
  conditions: Condition[];
}

// 任务类型
export interface Job {
  id: number;
  uuid: string;
  kind: JobKind;
  status: JobStatus;
  input: string;
  createTime: string;
  updateTime: string;
  detect?: DetectOptions;
  videoSegment?: VideoSegmentOptions;
  device: DeviceSpec;
  workflowId?: number;
  query?: string;
  resultFilter?: FilterCondition;
}

export interface JobSpec {
  id: number;
  uuid: string;
  kind: JobKind;
  status: JobStatus;
  input: string;
  createTime: string;
  updateTime: string;
  detect?: DetectOptions;
  videoSegment?: VideoSegmentOptions;
  device: DeviceSpec;
  workflowId?: number;
  query?: string;
  resultFilter?: FilterCondition;
}

export interface CreateJobRequest {
  uuid?: string;
  kind: JobKind;
  input: string;
  detect?: DetectOptions;
  videoSegment?: VideoSegmentOptions;
  deviceId: number;
  workflowId?: number;
  query?: string;
  resultFilter?: FilterCondition;
}

export interface UpdateJobRequest {
  kind?: JobKind;
  status?: JobStatus;
  input?: string;
  detect?: DetectOptions;
  videoSegment?: VideoSegmentOptions;
  device?: DeviceSpec;
  workflowId?: number;
  query?: string;
  resultFilter?: FilterCondition;
}

export interface CreateJobResponse {
  uuid: string;
  kind: JobKind;
  status: JobStatus;
}

export interface ListJobResponse {
  items: Job[];
  total: number;
}

// Detection box type for message
export interface DetectionBox {
  x: number;
  y: number;
  width: number;
  height: number;
  confidence: number;
  class: string;
}

// Workflow response type
export interface WorkflowResp {
  answer: string;
  [key: string]: any;
}

// 消息类型
export interface Message {
  id: number;
  jobId: number;
  timestamp: string;
  imagePath?: string;
  detectBoxes?: DetectionBox[];
  videoPath?: string;
  createTime: string;
  workflowResp?: WorkflowResp;
}

export interface MessageSpec {
  id: number;
  jobId: number;
  timestamp: string;
  imagePath?: string;
  detectBoxes?: DetectionBox[];
  videoPath?: string;
  createTime: string;
  workflowResp?: WorkflowResp;
}

export interface ListMessageResponse {
  items: Message[];
  total: number;
}

// 工作流类型
export interface Workflow {
  id: number;
  uuid: string;
  key: string;
  endpoint: string;
  name: string;
  timeout: number;
  createTime: string;
}

export interface WorkflowSpec {
  id: number;
  uuid: string;
  key: string;
  endpoint: string;
  name: string;
  timeout: number;
  createTime: string;
}

export interface CreateWorkflowRequest {
  name: string;
  endpoint: string;
  key: string;
  timeout?: number;
}

export interface UpdateWorkflowRequest {
  name?: string;
  endpoint?: string;
  key?: string;
  timeout?: number;
}

export interface CreateWorkflowResponse {
  id: number;
  name: string;
  uuid: string;
}

export interface ListWorkflowResponse {
  workflows: Workflow[];
  total: number;
}

export interface ListParams {
  start: number;
  limit: number;
  jobId?: number;
  alerted?: boolean;
}

export interface RouteParams {
  id?: string;
}

export interface JobStatsRequest {
  start?: string; // RFC3339 string
  end?: string;   // RFC3339 string
  window?: string; // e.g., '1m', '5m'
}

export interface TimeCount {
  time: string;
  count: number;
}

export interface LabelTimeCount {
  label: string;
  time: string;
  count: number;
}

export interface JobStatsResponse {
  messages: TimeCount[];
  labels?: LabelTimeCount[];
}