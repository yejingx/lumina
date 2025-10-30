import React, { useState, useEffect, useRef } from 'react';
import { Card, Descriptions, Button, Space, message, Modal, Tag, Drawer, Tabs, Spin } from 'antd';
import { EditOutlined, DeleteOutlined, ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { cameraApi } from '../../services/api';
import { formatDate } from '../../utils/helpers';
import type { CameraSpec, PreviewTask } from '../../types';
import CameraForm from './CameraForm';
import FlvPlayer from '../../components/FlvPlayer';

const CameraDetail: React.FC = () => {
  const [camera, setCamera] = useState<CameraSpec | null>(null);
  const [loading, setLoading] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [activeTab, setActiveTab] = useState<string>('detail');
  const [previewTask, setPreviewTask] = useState<PreviewTask | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const retryTimerRef = useRef<number | null>(null);
  const touchTimerRef = useRef<number | null>(null);
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const fetchCamera = async () => {
    if (!id) return;

    setLoading(true);
    try {
      const response = await cameraApi.get(parseInt(id));
      setCamera(response);
    } catch (error) {
      message.error('获取摄像头详情失败');
      console.error('Error fetching camera:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    setDrawerVisible(true);
  };

  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchCamera();
  };

  const startPreview = async () => {
    if (!id) return;
    setPreviewLoading(true);
    try {
      const task = await cameraApi.startPreview(parseInt(id));
      setPreviewTask(task);
    } catch (error) {
      message.error('开启预览失败');
      // eslint-disable-next-line no-console
      console.error('Error start preview:', error);
      // 如果拉取播放地址失败，5秒后重试
      scheduleRetry();
    } finally {
      setPreviewLoading(false);
    }
  };

  // 当播放失败时，隔5秒重试获取播放地址
  const scheduleRetry = () => {
    if (retryTimerRef.current) return; // 已有重试定时器时不重复设置
    const timer = window.setTimeout(async () => {
      try {
        await startPreview();
      } finally {
        // 重试后清除计时器，若仍失败会由错误回调再次触发
        retryTimerRef.current = null;
      }
    }, 5000);
    retryTimerRef.current = timer;
  };

  // 播放成功后才开始触摸心跳定时器
  const startTouchTimer = () => {
    if (touchTimerRef.current) return; // 已有定时器不重复设置
    const timer = window.setInterval(() => {
      touchPreview();
    }, 15000);
    touchTimerRef.current = timer;
  };

  const touchPreview = async () => {
    if (!id) return;
    try {
      await cameraApi.touchPreview(parseInt(id));
    } catch (error) {
      // 触摸失败不打断播放，仅记录
      // eslint-disable-next-line no-console
      console.warn('Touch preview failed:', error);
    }
  };

  useEffect(() => {
    // Tab 切换到预览时启动预览
    if (activeTab === 'preview') {
      startPreview();
      // 不在此处启动 touch 定时器，待播放成功后再启动
    }
    return () => {
      if (touchTimerRef.current) {
        window.clearInterval(touchTimerRef.current);
        touchTimerRef.current = null;
      }
      if (retryTimerRef.current) {
        window.clearTimeout(retryTimerRef.current);
        retryTimerRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, id]);

  useEffect(() => {
    fetchCamera();
  }, [id]);

  // 根据查询参数激活指定标签页
  useEffect(() => {
    const tab = searchParams.get('tab');
    if (tab === 'preview' || tab === 'detail') {
      setActiveTab(tab);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  const handleDelete = () => {
    if (!camera) return;

    Modal.confirm({
      title: '确认删除',
      content: `确定要删除摄像头 "${camera.name}" 吗？`,
      onOk: async () => {
        try {
          await cameraApi.delete(camera.id);
          message.success('删除成功');
          navigate('/cameras');
        } catch (error) {
          message.error('删除失败');
          console.error('Error deleting camera:', error);
        }
      },
    });
  };

  const getProtocolColor = (protocol: string) => {
    switch (protocol.toLowerCase()) {
      case 'rtsp':
        return 'blue';
      case 'rtmp':
        return 'volcano';
      default:
        return 'default';
    }
  };

  if (!camera && !loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px 0' }}>
          <p>摄像头不存在</p>
          <Button onClick={() => navigate('/camera')}>返回列表</Button>
          <Button onClick={() => navigate('/cameras')}>返回列表</Button>
        </div>
      </Card>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/cameras')}
          >
            返回列表
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => fetchCamera()}
          >
            刷新
          </Button>
        </Space>
        <Space>
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={handleEdit}
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
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'detail',
            label: '详情',
            children: (
              <Card title="摄像头详情" loading={loading}>
                {camera && (
                  <Descriptions column={2} bordered>
                    <Descriptions.Item label="UUID">{camera.uuid}</Descriptions.Item>
                    <Descriptions.Item label="名称">{camera.name}</Descriptions.Item>
                    <Descriptions.Item label="绑定设备">
                      {camera.bindDevice?.name || ''}
                    </Descriptions.Item>
                    <Descriptions.Item label="协议">
                      <Tag color={getProtocolColor(camera.protocol)}>
                        {camera.protocol.toUpperCase()}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="IP地址">{camera.ip}</Descriptions.Item>
                    <Descriptions.Item label="端口">{camera.port}</Descriptions.Item>
                    <Descriptions.Item label="路径" span={2}>
                      {camera.path || '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label="用户名">
                      {camera.username || '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label="密码">
                      {camera.password ? '••••••••' : '-'}
                    </Descriptions.Item>
                    <Descriptions.Item label="创建时间">
                      {formatDate(camera.createTime)}
                    </Descriptions.Item>
                    <Descriptions.Item label="更新时间">
                      {formatDate(camera.updateTime)}
                    </Descriptions.Item>
                  </Descriptions>
                )}
              </Card>
            ),
          },
          {
            key: 'preview',
            label: '预览',
            children: (
              <Card title="摄像头预览" extra={
                <Button onClick={startPreview} loading={previewLoading}>
                  重新获取播放地址
                </Button>
              }>
                {previewLoading && (
                  <div style={{ textAlign: 'center', padding: '24px' }}>
                    <Spin />
                  </div>
                )}
                {!previewLoading && previewTask?.previewAddr ? (
                  <FlvPlayer
                    url={previewTask.previewAddr}
                    onPlaying={() => {
                      // 播放成功，确保没有残留的重试定时器
                      if (retryTimerRef.current) {
                        window.clearTimeout(retryTimerRef.current);
                        retryTimerRef.current = null;
                      }
                      // 播放成功后再开始 touch 定时器
                      startTouchTimer();
                    }}
                    onError={() => {
                      // 播放错误，计划5秒后重试
                      scheduleRetry();
                    }}
                  />
                ) : (
                  <div style={{ textAlign: 'center', padding: '24px', color: '#999' }}>
                    点击上方按钮开始预览
                  </div>
                )}
              </Card>
            ),
          },
        ]}
      />

      <Drawer
        title="编辑摄像头"
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <CameraForm
          mode="drawer"
          cameraId={id}
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </div>
  );
};

export default CameraDetail;