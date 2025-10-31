import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Modal,
  message,
  Tabs,
  Tag,
  Typography,
  Spin,
  Table,
  Input,
  Collapse,
  Drawer,
  Image,
  Row,
  Col,
  DatePicker,
  Select,
  Checkbox,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  EyeOutlined,
  FileImageOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { jobApi, messageApi, deviceApi, workflowApi } from '../../services/api';
import type { Job, Message, Device, Workflow, ListParams, JobStatsResponse, JobStatsRequest, CameraSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig, isOlderThanMinutes } from '../../utils/helpers';
import { JOB_STATUS_MAP, JOB_KIND_MAP, MESSAGE_TYPE_MAP, DEFAULT_PAGE_SIZE } from '../../utils/constants';
import JobForm from './JobForm';
import VideoThumbnail from '../../components/VideoThumbnail';
import { Line } from '@ant-design/plots';

const { Title, Text } = Typography;
const { Search } = Input;
const { Panel } = Collapse;



const JobDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [job, setJob] = useState<Job | null>(null);
  const [device, setDevice] = useState<Device | null>(null);
  const [workflow, setWorkflow] = useState<Workflow | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [messageLoading, setMessageLoading] = useState(false);
  const [messageParams, setMessageParams] = useState<ListParams>({
    start: 0,
    limit: DEFAULT_PAGE_SIZE,
  });
  const [messageTotal, setMessageTotal] = useState(0);
  const [activeTab, setActiveTab] = useState('basic');
  const [messagesCurrent, setMessagesCurrent] = useState(1);
  const [messagesTotal, setMessagesTotal] = useState(0);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [editDrawerVisible, setEditDrawerVisible] = useState(false);
  const [onlyAlerted, setOnlyAlerted] = useState(true);

  // Navigate to device detail when clicking host name
  const handleViewDevice = (d?: Device) => {
    const id = d?.id;
    if (id !== undefined && id !== null) {
      navigate(`/devices/${id}`);
    }
  };

  // Navigate to camera detail when clicking camera
  const handleViewCamera = (c?: CameraSpec) => {
    const id = c?.id;
    if (id !== undefined && id !== null) {
      navigate(`/cameras/${id}`);
    }
  };

  // Navigate to workflow detail when clicking workflow
  const handleViewWorkflow = (w?: Workflow) => {
    const id = w?.id;
    if (id !== undefined && id !== null) {
      navigate(`/workflows/${id}`);
    }
  };

  // 统计数据状态
  const [stats, setStats] = useState<JobStatsResponse | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const [statsParams, setStatsParams] = useState<JobStatsRequest>({ window: '5m' });

  // 统计区间选择（非必需，用户选择时才更新）
  const handleRangeChange = (values: any) => {
    if (!values || values.length !== 2) {
      setStatsParams((prev) => ({ ...prev, start: undefined, end: undefined }));
      return;
    }
    const [start, end] = values;
    setStatsParams((prev) => ({
      ...prev,
      start: start?.toDate()?.toISOString(),
      end: end?.toDate()?.toISOString(),
    }));
  };

  const handleWindowChange = (value: string) => {
    setStatsParams((prev) => ({ ...prev, window: value }));
  };

  // 获取任务详情
  const fetchJobDetail = async () => {
    setLoading(true);
    try {
      const jobId = parseInt(id!);
      const jobData = await jobApi.get(jobId);
      setJob(jobData);

      // 获取关联的设备信息 - support both device object and device_id
      if (jobData.device) {
        setDevice(jobData.device);
      } else if ((jobData as any).device_id) {
        try {
          const deviceData = await deviceApi.get((jobData as any).device_id);
          setDevice(deviceData);
        } catch (error) {
          console.warn('Failed to fetch device:', error);
        }
      }

      if ((jobData as any).workflow) {
        setWorkflow((jobData as any).workflow);
      }
    } catch (error) {
      handleApiError(error, '获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取任务消息
  const fetchJobMessages = async () => {
    if (!id || !job) return;
    setMessagesLoading(true);
    try {
      const params: ListParams = {
        start: (messagesCurrent - 1) * DEFAULT_PAGE_SIZE,
        limit: DEFAULT_PAGE_SIZE,
        jobId: job.id,
        alerted: onlyAlerted ? true : undefined,
      };
      const response = await messageApi.list(params);
      setMessages(response.items);
      setMessagesTotal(response.total);
    } catch (error) {
      handleApiError(error, '获取任务消息失败');
    } finally {
      setMessagesLoading(false);
    }
  };

  useEffect(() => {
    fetchJobDetail();
  }, [id]);

  useEffect(() => {
    if (activeTab === 'messages') {
      fetchJobMessages();
    }
  }, [activeTab, messagesCurrent, id, job, onlyAlerted]);

  // 获取任务统计
  const fetchJobStats = async () => {
    if (!id || !job) return;
    setStatsLoading(true);
    try {
      const jobId = job.id;
      const data = await jobApi.stats(jobId, statsParams);
      setStats(data);
    } catch (error) {
      handleApiError(error, '获取任务统计失败');
    } finally {
      setStatsLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'stats') {
      fetchJobStats();
    }
  }, [activeTab, id, job, statsParams]);

  // 处理任务操作
  const handleJobAction = async (action: string) => {
    if (!job) return;

    try {
      // Use uuid if available, fallback to id for backward compatibility
      const identifier = job.id;
      if (identifier) {
        if (action === 'start') {
          await jobApi.start(identifier);
          message.success('任务启动成功');
        } else if (action === 'pause' || action === 'stop') {
          await jobApi.stop(identifier);
          message.success('任务停止成功');
        }
        fetchJobDetail();
      }
    } catch (error) {
      handleApiError(error, '操作失败');
    }
  };

  // 处理删除任务
  const handleDelete = () => {
    if (!job) return;

    Modal.confirm({
      ...getDeleteConfirmConfig(`删除任务 "${job.uuid}"`),
      onOk: async () => {
        try {
          // Use uuid if available, fallback to id for backward compatibility
          const identifier = job.uuid || job.id?.toString();
          if (identifier) {
            await jobApi.delete(identifier as any);
            message.success('删除成功');
            navigate('/jobs');
          }
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // Render thumbnail for image or video
  const renderThumbnail = (message: Message) => {
    if (message.imagePath) {
      return (
        <Image
          width={60}
          height={40}
          src={message.imagePath}
          alt="图片缩略图"
          style={{ objectFit: 'cover', borderRadius: 4 }}
          preview={{
            mask: <EyeOutlined />
          }}
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

  // 消息表格列配置
  const messageColumns: ColumnsType<Message> = [
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
      width: 400,
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
      render: (alerted: boolean) => (alerted ? '是' : '否'),
    },
    {
      title: '时间戳',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => formatDate(timestamp),
    },
  ];



  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!job) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Text type="secondary">任务不存在</Text>
        </div>
      </Card>
    );
  }

  const isUnknown = isOlderThanMinutes(job?.device?.lastPingTime, 10);
  const statusInfo = isUnknown
    ? { text: '未知', color: 'default' as const }
    : (JOB_STATUS_MAP[job.status as keyof typeof JOB_STATUS_MAP] || { text: job.status, color: 'default' });
  const kindInfo = JOB_KIND_MAP[(job.kind) as keyof typeof JOB_KIND_MAP];

  const tabItems = [
    {
      key: 'basic',
      label: '基本信息',
      children: (
        <div>
          <Descriptions column={2} bordered>
            <Descriptions.Item label="UUID">{job.uuid || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="类型">
              {kindInfo ? (
                <Tag color={kindInfo.color}>{kindInfo.text}</Tag>
              ) : (
                job.kind || '-'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="摄像头">
              {job?.camera ? (
                <Button
                  type="link"
                  size="small"
                  onClick={() => handleViewCamera(job.camera)}
                  style={{ padding: 0 }}
                >
                  {job.camera.name || job.camera.uuid || '-'}
                </Button>
              ) : (
                '-'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="运行主机">
              {job?.device ? (
                <Button
                  type="link"
                  size="small"
                  onClick={() => handleViewDevice(job.device)}
                  style={{ padding: 0 }}
                >
                  {job.device.name || job.device.uuid || '-'}
                </Button>
              ) : (
                '-'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="工作流">
              {job?.workflow ? (
                <Button
                  type="link"
                  size="small"
                  onClick={() => handleViewWorkflow(job.workflow)}
                  style={{ padding: 0 }}
                >
                  {job.workflow.name || (job as any).workflow?.uuid || '-'}
                </Button>
              ) : (
                '-'
              )}
            </Descriptions.Item>

            <Descriptions.Item label="创建时间">
              {formatDate(job.createTime || (job as any).created_at)}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {formatDate(job.updateTime || (job as any).updated_at)}
            </Descriptions.Item>
            <Descriptions.Item label="开始时间">
              {formatDate((job as any).started_at) || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="完成时间">
              {formatDate((job as any).finished_at) || '-'}
            </Descriptions.Item>
          </Descriptions>

          {/* Display detect options */}
          {job.detect && (
            <Card title="检测参数" style={{ marginTop: 16 }} size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="模型名称">
                  {job.detect.modelName || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="标签">
                  {job.detect.labels || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="置信度阈值">
                  {job.detect.confThreshold !== undefined ? job.detect.confThreshold : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="IoU阈值">
                  {job.detect.iouThreshold !== undefined ? job.detect.iouThreshold : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="检测间隔">
                  {job.detect.interval !== undefined ? `${job.detect.interval} ms` : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="触发次数">
                  {job.detect.triggerCount !== undefined ? job.detect.triggerCount : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="触发间隔">
                  {job.detect.triggerInterval !== undefined ? `${job.detect.triggerInterval} 秒` : '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          {/* Display video segment options */}
          {job.videoSegment && (
            <Card title="视频分割参数" style={{ marginTop: 16 }} size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="分割间隔">
                  {job.videoSegment.interval !== undefined ? `${job.videoSegment.interval} 秒` : '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          {/* 结果过滤卡片已移除，使用查询设置下方的单行显示 */}
        </div>
      ),
    },
    {
      key: 'messages',
      label: '消息记录',
      children: (
        <div>
          <div style={{ marginBottom: 16 }}>
            <Space style={{ marginBottom: 16 }}>
              <Button
                icon={<ReloadOutlined />}
                onClick={fetchJobMessages}
              >
                刷新
              </Button>
              <Checkbox
                checked={onlyAlerted}
                onChange={(e) => setOnlyAlerted(e.target.checked)}
                style={{ marginLeft: 20 }}
              >
                只看已告警消息
              </Checkbox>
            </Space>
            <Space style={{ float: 'right' }}>
              <Search
                placeholder="搜索工作流回答"
                allowClear
                style={{ width: 200 }}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onSearch={fetchJobMessages}
              />
            </Space>
          </div>
          <Table
            columns={messageColumns}
            dataSource={messages}
            rowKey="id"
            loading={messagesLoading}
            pagination={{
              current: messagesCurrent,
              pageSize: DEFAULT_PAGE_SIZE,
              total: messagesTotal,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) =>
                `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
              onChange: (page) => {
                setMessagesCurrent(page);
              },
            }}
          />
        </div>
      ),
    },
    {
      key: 'stats',
      label: '统计',
      children: (
        <div>
          <Space style={{ marginBottom: 16 }}>
            <Button icon={<ReloadOutlined />} onClick={fetchJobStats}>刷新</Button>
            <DatePicker.RangePicker
              showTime
              onChange={handleRangeChange}
              style={{ minWidth: 280 }}
            />
            <Select
              value={statsParams.window || '5m'}
              onChange={handleWindowChange}
              style={{ width: 120 }}
              options={[
                { value: '1m', label: '1分钟' },
                { value: '5m', label: '5分钟' },
                { value: '15m', label: '15分钟' },
                { value: '1h', label: '1小时' },
              ]}
            />
          </Space>

          {statsLoading ? (
            <Spin />
          ) : (
            <Row gutter={16}>
              <Col span={12}>
                <Card title="消息数量趋势" style={{ marginBottom: 16 }}>
                  <Line
                    data={(stats?.messages || []).map((d: any) => ({ time: d.time, count: d.count }))}
                    xField="time"
                    yField="count"
                    xAxis={{ type: 'time' }}
                    smooth
                  />
                </Card>
              </Col>
              {job?.kind === 'detect' && (
                <Col span={12}>
                  <Card title="标签数量趋势">
                    <Line
                      data={(stats?.labels || []).map((d: any) => ({ time: d.time, count: d.count, label: d.label }))}
                      xField="time"
                      yField="count"
                      seriesField="label"
                      xAxis={{ type: 'time' }}
                      smooth
                    />
                  </Card>
                </Col>
              )}
            </Row>
          )}
        </div>
      ),
    },
  ];

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/jobs')}
            >
              返回列表
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchJobDetail}
            >
              刷新
            </Button>
          </Space>
          <Space>
            {!job.enabled && (
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={() => handleJobAction('start')}
                style={{ backgroundColor: '#52c41a', borderColor: '#52c41a' }}
              >
                启动
              </Button>
            )}
            {job.enabled && (
              <Button
                type="default"
                icon={<StopOutlined />}
                onClick={() => handleJobAction('stop')}
                danger
              >
                停止
              </Button>
            )}
            <Button
              icon={<EditOutlined />}
              onClick={() => setEditDrawerVisible(true)}
            >
              编辑
            </Button>
            <Button
              danger
              icon={<DeleteOutlined />}
              onClick={handleDelete}
            >
              删除
            </Button>
          </Space>
        </Space>
      </Card>

      <Card title={`任务详情 - ${`${job.uuid}`}`}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
        />
      </Card>

      {/* Edit Job Drawer */}
      <Drawer
        title="编辑任务"
        width={600}
        open={editDrawerVisible}
        onClose={() => setEditDrawerVisible(false)}
        destroyOnClose
      >
        <JobForm
          job={job}
          onSubmit={() => {
            setEditDrawerVisible(false);
            fetchJobDetail(); // Refresh job details after edit
          }}
          onCancel={() => setEditDrawerVisible(false)}
        />
      </Drawer>
    </div>
  );
};

export default JobDetail;