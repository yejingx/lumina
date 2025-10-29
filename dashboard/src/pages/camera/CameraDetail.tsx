import React, { useState, useEffect } from 'react';
import { Card, Descriptions, Button, Space, message, Modal, Tag, Drawer } from 'antd';
import { EditOutlined, DeleteOutlined, ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { cameraApi } from '../../services/api';
import { formatDate } from '../../utils/helpers';
import type { CameraSpec } from '../../types';
import CameraForm from './CameraForm';

const CameraDetail: React.FC = () => {
  const [camera, setCamera] = useState<CameraSpec | null>(null);
  const [loading, setLoading] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

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

  useEffect(() => {
    fetchCamera();
  }, [id]);

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

      <Card title="摄像头详情" loading={loading}>
        {camera && (
          <Descriptions column={2} bordered>
            <Descriptions.Item label="ID">{camera.id}</Descriptions.Item>
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