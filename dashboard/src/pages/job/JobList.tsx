import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  message,
  Card,
  Input,
  Select,
  Drawer,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { jobApi } from '../../services/api';
import type { Job, ListParams, CameraSpec, DeviceSpec, WorkflowSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig, isOlderThanMinutes } from '../../utils/helpers';
import { JOB_STATUS_MAP, JOB_KIND_MAP, DEFAULT_PAGE_SIZE } from '../../utils/constants';
import JobForm from './JobForm';

const { Search } = Input;
const { Option } = Select;

const JobList: React.FC = () => {
  const navigate = useNavigate();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingJob, setEditingJob] = useState<Job | null>(null);

  // 获取任务列表
  const fetchJobs = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await jobApi.list(params);
      setJobs(response.items || []);
      setTotal(response.total || 0);
    } catch (error) {
      console.error('Failed to fetch jobs:', error);
      message.error('获取任务列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchJobs();
  }, [current, pageSize]);

  // 处理创建任务
  const handleCreate = () => {
    setEditingJob(null);
    setDrawerVisible(true);
  };

  // 处理编辑任务
  const handleEdit = (job: Job) => {
    setEditingJob(job);
    setDrawerVisible(true);
  };

  // 处理删除任务
  const handleDelete = (job: Job) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除任务 "${job.uuid}"`),
      onOk: async () => {
        try {
          // Use uuid if available, fallback to id for backward compatibility
          const identifier = job.id;
          if (identifier) {
            await jobApi.delete(identifier);
            message.success('删除成功');
            fetchJobs();
          }
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理查看详情
  const handleView = (job: Job) => {
    const identifier = job.id;
    navigate(`/jobs/${identifier}`);
  };

  // 处理查看摄像头详情
  const handleViewCamera = (camera: CameraSpec | undefined) => {
    const id = camera?.id;
    if (id !== undefined && id !== null) {
      navigate(`/cameras/${id}`);
    }
  };

  // 处理查看设备详情
  const handleViewDevice = (device: DeviceSpec | undefined) => {
    const id = device?.id;
    if (id !== undefined && id !== null) {
      navigate(`/devices/${id}`);
    }
  };

  // 处理查看工作流详情
  const handleViewWorkflow = (workflow: WorkflowSpec | undefined) => {
    const id = workflow?.id;
    if (id !== undefined && id !== null) {
      navigate(`/workflows/${id}`);
    }
  };

  // 处理启动任务
  const handleStart = async (job: Job) => {
    try {
      const identifier = job.id;
      if (identifier) {
        await jobApi.start(identifier as any);
        message.success('任务启动成功');
        fetchJobs();
      }
    } catch (error) {
      handleApiError(error, '启动任务失败');
    }
  };

  // 处理停止任务
  const handleStop = async (job: Job) => {
    try {
      const identifier = job.id;
      if (identifier) {
        await jobApi.stop(identifier as any);
        message.success('任务停止成功');
        fetchJobs();
      }
    } catch (error) {
      handleApiError(error, '停止任务失败');
    }
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchJobs();
  };

  // 表格列配置
  const columns: ColumnsType<Job> = [
    {
      title: 'UUID',
      dataIndex: 'uuid',
      key: 'uuid',
      width: 130,
      ellipsis: true,
      render: (uuid: string, record: Job) => (
        <Button
          type="link"
          size="small"
          onClick={() => handleView(record)}
          style={{ padding: 0 }}
        >
          {uuid || record.id || '-'}
        </Button>
      ),
    },
    {
      title: '类型',
      dataIndex: 'kind',
      key: 'kind',
      width: 80,
      render: (kind: string, record: Job) => {
        const jobKind = kind;
        if (jobKind) {
          const kindInfo = JOB_KIND_MAP[jobKind as keyof typeof JOB_KIND_MAP];
          if (kindInfo) {
            return <Tag color={kindInfo.color}>{kindInfo.text}</Tag>;
          }
        }
        return jobKind || '-';
      },
    },
    {
      title: '摄像头',
      dataIndex: 'camera',
      key: 'camera',
      width: 100,
      ellipsis: true,
      render: (camera: any, record: Job) => {
        if (!camera) return '-';
        return (
          <Button
            type="link"
            size="small"
            onClick={() => handleViewCamera(camera)}
            style={{ padding: 0 }}
          >
            {camera.name || camera.uuid || '-'}
          </Button>
        );
      },
    },
    {
      title: '运行主机',
      dataIndex: 'device',
      key: 'device',
      width: 100,
      ellipsis: true,
      render: (device: any, record: Job) => {
        if (!device) return '-';
        return (
          <Button
            type="link"
            size="small"
            onClick={() => handleViewDevice(device)}
            style={{ padding: 0 }}
          >
            {device.name || device.uuid || '-'}
          </Button>
        );
      },
    },
    {
      title: '工作流',
      dataIndex: 'workflow',
      key: 'workflow',
      width: 100,
      ellipsis: true,
      render: (_: any, record: Job) => {
        const wf = record.workflow;
        if (!wf) return '-';
        return (
          <Button
            type="link"
            size="small"
            onClick={() => handleViewWorkflow(wf)}
            style={{ padding: 0 }}
          >
            {wf.name || wf.uuid || '-'}
          </Button>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 50,
      render: (status: string, record: Job) => {
        const lastPing = record?.device?.lastPingTime;
        const isUnknown = isOlderThanMinutes(lastPing, 10);
        if (isUnknown) {
          return <Tag color="default">未知</Tag>;
        }
        const statusInfo = JOB_STATUS_MAP[status as keyof typeof JOB_STATUS_MAP] || {
          text: status,
          color: 'default',
        };
        return <Tag color={statusInfo.color}>{statusInfo.text}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 120,
      render: (createTime: string, record: Job) => formatDate(createTime),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleView(record)}
            title="查看详情"
          />
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            title="编辑"
          />
          {!record.enabled && (
            <Button
              type="text"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleStart(record)}
              title="启动"
              style={{ color: '#52c41a' }}
            />
          )}
          {record.enabled && (
            <Button
              type="text"
              size="small"
              icon={<PauseCircleOutlined />}
              onClick={() => handleStop(record)}
              title="停止"
              style={{ color: '#ff4d4f' }}
            />
          )}
          <Button
            type="text"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
            title="删除"
          />
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <div style={{ marginBottom: 16 }}>
        <Space style={{ marginBottom: 16 }}>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            创建任务
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchJobs}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Search
            placeholder="搜索任务名称"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchJobs}
          />
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 120 }}
            value={statusFilter}
            onChange={(value) => setStatusFilter(value)}
          >
            {Object.entries(JOB_STATUS_MAP).map(([key, value]) => (
              <Option key={key} value={key}>
                {value.text}
              </Option>
            ))}
          </Select>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={jobs}
        rowKey={(record) => record.uuid || record.id?.toString() || ''}
        loading={loading}
        pagination={{
          current,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) =>
            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          onChange: (page, size) => {
            setCurrent(page);
            setPageSize(size || DEFAULT_PAGE_SIZE);
          },
        }}
      />

      <Drawer
        title={editingJob ? '编辑任务' : '创建任务'}
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <JobForm
          job={editingJob}
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </Card>
  );
};

export default JobList;