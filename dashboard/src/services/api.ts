import axios from 'axios';
import type {
  // User types
  ListUsersResponse,
  CreateUserRequest,
  CreateUserResponse,
  User,

  // Device types
  ListDeviceResponse,
  RegisterRequest,
  RegisterResponse,
  DeviceSpec,

  // Access Token types
  ListAccessTokenResponse,
  CreateAccessTokenRequest,
  CreateAccessTokenResponse,
  AccessTokenSpec,

  // Job types
  ListJobResponse,
  CreateJobRequest,
  CreateJobResponse,
  UpdateJobRequest,
  JobSpec,

  // Message types
  ListMessageResponse,
  MessageSpec,

  // Workflow types
  ListWorkflowResponse,
  CreateWorkflowRequest,
  CreateWorkflowResponse,
  WorkflowSpec,

  ListParams,
} from '../types';

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证 token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// 用户 API
export const userApi = {
  // 获取用户列表
  list: (params: ListParams): Promise<ListUsersResponse> =>
    api.get('/admin/users', { params }),

  // 创建用户
  create: (data: CreateUserRequest): Promise<CreateUserResponse> =>
    api.post('/admin/users', data),

  // 删除用户
  delete: (userId: number): Promise<void> =>
    api.delete(`/admin/user/${userId}`),
};

// 设备 API
export const deviceApi = {
  // 获取设备列表
  list: (params: ListParams): Promise<ListDeviceResponse> =>
    api.get('/device', { params }),

  // 获取设备详情
  get: (deviceId: number): Promise<DeviceSpec> =>
    api.get(`/device/${deviceId}`),

  // 注册设备
  register: (data: RegisterRequest): Promise<RegisterResponse> =>
    api.post('/agent/register', data),

  // 注销设备
  unregister: (): Promise<void> =>
    api.post('/agent/unregister'),

  // 删除设备
  delete: (deviceId: number): Promise<void> =>
    api.delete(`/device/${deviceId}`),
};

// 接入凭证 API
export const accessTokenApi = {
  // 获取接入凭证列表
  list: (params: ListParams): Promise<ListAccessTokenResponse> =>
    api.get('/access-token', { params }),

  // 获取接入凭证详情
  get: (tokenId: number): Promise<AccessTokenSpec> =>
    api.get(`/access-token/${tokenId}`),

  // 创建接入凭证
  create: (data: CreateAccessTokenRequest): Promise<CreateAccessTokenResponse> =>
    api.post('/access-token', data),

  // 删除接入凭证
  delete: (tokenId: number): Promise<void> =>
    api.delete(`/access-token/${tokenId}`),
};

// 任务 API
export const jobApi = {
  // 获取任务列表
  list: (params: ListParams): Promise<ListJobResponse> =>
    api.get('/job', { params }),

  // 获取任务详情
  get: (jobId: number): Promise<JobSpec> =>
    api.get(`/job/${jobId}`),

  // 创建任务
  create: (data: CreateJobRequest): Promise<CreateJobResponse> =>
    api.post('/job', data),

  // 更新任务
  update: (jobId: number, data: Partial<UpdateJobRequest>): Promise<JobSpec> =>
    api.put(`/job/${jobId}`, data),

  // 删除任务
  delete: (jobId: number): Promise<void> =>
    api.delete(`/job/${jobId}`),

  // 启动任务
  start: (jobId: string | number): Promise<void> =>
    api.put(`/job/${jobId}/start`),

  // 停止任务
  stop: (jobId: string | number): Promise<void> =>
    api.put(`/job/${jobId}/stop`),
};

// 消息 API
export const messageApi = {
  // 获取消息列表
  list: (params: ListParams & { jobId?: number }): Promise<ListMessageResponse> =>
    api.get('/message', { params }),

  // 获取消息详情
  get: (messageId: number): Promise<MessageSpec> =>
    api.get(`/message/${messageId}`),

  // 删除消息
  delete: (messageId: number): Promise<void> =>
    api.delete(`/message/${messageId}`),
};

// 工作流 API
export const workflowApi = {
  // 获取工作流列表
  list: (params: ListParams): Promise<ListWorkflowResponse> =>
    api.get('/workflow', { params }),

  // 获取工作流详情
  get: (workflowId: number): Promise<WorkflowSpec> =>
    api.get(`/workflow/${workflowId}`),

  // 创建工作流
  create: (data: CreateWorkflowRequest): Promise<CreateWorkflowResponse> =>
    api.post('/workflow', data),

  // 更新工作流
  update: (workflowId: number, data: Partial<CreateWorkflowRequest>): Promise<WorkflowSpec> =>
    api.put(`/workflow/${workflowId}`, data),

  // 删除工作流
  delete: (workflowId: number): Promise<void> =>
    api.delete(`/workflow/${workflowId}`),
};

export default api;