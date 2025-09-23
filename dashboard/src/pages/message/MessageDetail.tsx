import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Button,
  Space,
  Modal,
  message,
  Typography,
  Spin,
  Image,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  FileImageOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { messageApi, jobApi } from '../../services/api';
import type { MessageSpec, JobSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import VideoThumbnail from '../../components/VideoThumbnail';

const { Title, Text, Paragraph } = Typography;

const MessageDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [messageDetail, setMessageDetail] = useState<MessageSpec | null>(null);
  const [job, setJob] = useState<JobSpec | null>(null);
  const [loading, setLoading] = useState(false);

  // 获取消息详情
  const fetchMessageDetail = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const messageData = await messageApi.get(parseInt(id));
      setMessageDetail(messageData);

      // 获取关联的任务信息
      if (messageData.jobId) {
        try {
          const jobData = await jobApi.get(messageData.jobId);
          setJob(jobData);
        } catch (error) {
          console.error('获取任务信息失败:', error);
        }
      }
    } catch (error) {
      handleApiError(error, '获取消息详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMessageDetail();
  }, [id]);

  // 处理删除消息
  const handleDelete = () => {
    if (!messageDetail) return;

    Modal.confirm({
      ...getDeleteConfirmConfig(`删除消息 ID: ${messageDetail.id}`),
      onOk: async () => {
        try {
          await messageApi.delete(messageDetail.id);
          message.success('删除成功');
          navigate('/messages');
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 渲染媒体内容
  const renderMedia = () => {
    if (!messageDetail) return null;

    if (messageDetail.imagePath) {
      return (
        <div>
          <div style={{ marginTop: 8 }}>
            <Image
              width={300}
              src={messageDetail.imagePath}
              alt="消息图片"
              style={{ borderRadius: 8 }}
            />
          </div>
        </div>
      );
    }

    if (messageDetail.videoPath) {
      return (
        <div>
          <div style={{ marginTop: 8 }}>
            <VideoThumbnail
              videoUrl={messageDetail.videoPath}
              width={300}
              height={200}
              enablePreview={true}
              title="点击播放视频"
              style={{ borderRadius: 8 }}
            />
          </div>
        </div>
      );
    }

    return null;
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!messageDetail) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Text type="secondary">消息不存在</Text>
        </div>
      </Card>
    );
  }

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/messages')}
              style={{ marginRight: 16 }}
            >
              返回
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              消息详情
            </Title>
          </div>
          <Space>
            <Button
              danger
              icon={<DeleteOutlined />}
              onClick={handleDelete}
            >
              删除
            </Button>
          </Space>
        </div>
      </Card>

      <Card>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="关联任务">
            {job ? (
              <Button
                type="link"
                onClick={() => navigate(`/jobs/${job.id}`)}
                style={{ padding: 0 }}
              >
                {job.uuid}
              </Button>
            ) : (
              messageDetail.jobId ? `任务 ${messageDetail.jobId}` : '-'
            )}
          </Descriptions.Item>
          <Descriptions.Item label="时间戳">
            {formatDate(messageDetail.timestamp)}
          </Descriptions.Item>
          {messageDetail.workflowResp?.answer && (
            <Descriptions.Item label="工作流回答" span={2}>
              <Paragraph
                copyable
                style={{
                  backgroundColor: '#f5f5f5',
                  padding: '12px',
                  borderRadius: '6px',
                  marginBottom: 0,
                  whiteSpace: 'pre-wrap',
                }}
              >
                {messageDetail.workflowResp.answer}
              </Paragraph>
            </Descriptions.Item>
          )}
          {messageDetail.detectBoxes && messageDetail.detectBoxes.length > 0 && (
            <Descriptions.Item label="检测框" span={2}>
              <div style={{
                backgroundColor: '#f5f5f5',
                padding: '12px',
                borderRadius: '6px',
                marginBottom: 0,
              }}>
                {messageDetail.detectBoxes.map((box, index) => (
                  <div key={index} style={{ marginBottom: 8 }}>
                    <Text strong>{box.class}</Text> - 
                    置信度: {(box.confidence * 100).toFixed(1)}%, 
                    位置: ({box.x}, {box.y}, {box.width}, {box.height})
                  </div>
                ))}
              </div>
            </Descriptions.Item>
          )}
          {(messageDetail.imagePath || messageDetail.videoPath) && (
            <Descriptions.Item label="媒体内容" span={2}>
              {renderMedia()}
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>
    </div>
  );
};

export default MessageDetail;