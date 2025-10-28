import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  message,
  Card,
  Input,
  Select,
  Drawer,
  Tag,
  Typography,
  Image,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  FileImageOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { messageApi, jobApi } from '../../services/api';
import type { Message, Job, ListParams } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import { DEFAULT_PAGE_SIZE } from '../../utils/constants';
import VideoThumbnail from '../../components/VideoThumbnail';

const { Search } = Input;
const { Option } = Select;
const { Text } = Typography;

const MessageList: React.FC = () => {
  const navigate = useNavigate();
  const [messages, setMessages] = useState<Message[]>([]);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  const [jobFilter, setJobFilter] = useState<number | undefined>();
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 获取任务列表
  const fetchJobs = async () => {
    try {
      const response = await jobApi.list({ start: 0, limit: 10 });
      setJobs(response.items || []);
    } catch (error) {
      console.error('获取任务列表失败:', error);
    }
  };

  // 获取消息列表
  const fetchMessages = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await messageApi.list(params);
      setMessages(response.items || []);
      setTotal(response.total || 0);
    } catch (error) {
      handleApiError(error, '获取消息列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchJobs();
  }, []);

  useEffect(() => {
    fetchMessages();
  }, [current, pageSize]);

  // 处理创建消息
  const handleCreate = () => {
    setDrawerVisible(true);
  };

  // 处理删除消息
  const handleDelete = (msg: Message) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除消息 ID: ${msg.id}`),
      onOk: async () => {
        try {
          await messageApi.delete(msg.id);
          message.success('删除成功');
          fetchMessages();
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理查看详情
  const handleView = (msg: Message) => {
    navigate(`/messages/${msg.id}`);
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchMessages();
  };

  // 获取任务名称
  const getJobName = (jobId: number) => {
    const job = jobs.find(j => j.id === jobId);
    return job ? job.uuid : `任务 ${jobId}`;
  };

  // Render thumbnail for image or video
  const renderThumbnail = (message: Message) => {
    if (message.imagePath) {
      return (
        <Image
          width={60}
          height={40}
          src={message.imagePath}
          style={{ objectFit: 'cover', borderRadius: 4 }}
          fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3Ik1RnG4W+FgYxN"
        />
      );
    } else if (message.videoPath) {
      return (
        <VideoThumbnail
          videoUrl={message.videoPath}
          width={60}
          height={40}
          enablePreview={true}
          title="点击播放视频"
        />
      );
    } else {
      return (
        <div style={{
          width: 60,
          height: 40,
          backgroundColor: '#f5f5f5',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: 4
        }}>
          <FileImageOutlined style={{ fontSize: 16, color: '#d9d9d9' }} />
        </div>
      );
    }
  };

  // 表格列配置
  const columns: ColumnsType<Message> = [
    {
      title: '任务ID',
      dataIndex: 'jobId',
      key: 'jobId',
      width: 180,
      render: (jobId: number) => (
        <Text>{getJobName(jobId)}</Text>
      ),
    },
    {
      title: '缩略图',
      key: 'thumbnail',
      width: 80,
      render: (_, record) => renderThumbnail(record),
    },
    {
      title: '工作流回答',
      dataIndex: 'workflowResp',
      key: 'workflowResp',
      width: 350,
      render: (workflowResp: any) => (
        <pre style={{
          maxHeight: '100px',
          overflow: 'auto',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          backgroundColor: '#f5f5f5',
          padding: '8px',
          borderRadius: '4px',
          fontSize: '12px',
          fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
          margin: 0,
          maxWidth: '450px'
        }}>
          {workflowResp?.answer || '-'}
        </pre>
      ),
    },
    {
      title: '已告警',
      dataIndex: 'alerted',
      key: 'alerted',
      width: 100,
      render: (alerted?: boolean) => (alerted ? '是' : '否'),
    },
    {
      title: '时间戳',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
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
            icon={<ReloadOutlined />}
            onClick={fetchMessages}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Select
            placeholder="筛选任务"
            allowClear
            style={{ width: 150 }}
            value={jobFilter}
            onChange={setJobFilter}
          >
            {jobs.map(job => (
              <Option key={job.id} value={job.id}>
                {job.uuid}
              </Option>
            ))}
          </Select>
          <Search
            placeholder="搜索工作流回答"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchMessages}
          />
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={messages}
        rowKey="id"
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

    </Card>
  );
};

export default MessageList;