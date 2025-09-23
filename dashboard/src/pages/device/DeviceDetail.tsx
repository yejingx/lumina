import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Button,
  Space,
  message,
  Spin,
  Tag,
  Drawer,
  Modal,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CopyOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { deviceApi } from '../../services/api';
import type { DeviceSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig, copyToClipboard } from '../../utils/helpers';
import DeviceForm from './DeviceForm';

const { Text } = Typography;

const DeviceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [device, setDevice] = useState<DeviceSpec | null>(null);
  const [loading, setLoading] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [tokenVisible, setTokenVisible] = useState(false);

  // 获取设备详情
  const fetchDeviceDetail = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const deviceId = parseInt(id, 10);
      const response = await deviceApi.get(deviceId);
      setDevice(response);
    } catch (error) {
      handleApiError(error, '获取设备详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDeviceDetail();
  }, [id]);

  // 处理编辑
  const handleEdit = () => {
    setDrawerVisible(true);
  };

  // 处理删除
  const handleDelete = () => {
    if (!device) return;
    
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除设备 "${device.uuid}"`),
      onOk: async () => {
        try {
          await deviceApi.delete(device.id);
          message.success('删除成功');
          navigate('/devices');
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理复制令牌
  const handleCopyToken = () => {
    if (device) {
      copyToClipboard(device.token);
    }
  };

  // 处理复制UUID
  const handleCopyUuid = () => {
    if (device) {
      copyToClipboard(device.uuid);
    }
  };

  // 切换令牌显示状态
  const toggleTokenVisibility = () => {
    setTokenVisible(!tokenVisible);
  };

  // 格式化令牌显示
  const formatToken = (token: string) => {
    return tokenVisible ? token : '•'.repeat(Math.min(token.length, 20));
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchDeviceDetail();
  };

  // 返回列表
  const handleBack = () => {
    navigate('/devices');
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!device) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <p>设备不存在或已被删除</p>
          <Button type="primary" onClick={handleBack}>
            返回列表
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
            >
              返回列表
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchDeviceDetail}
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
        </Space>
      </Card>

      <Card title={`设备详情 - ${device.uuid}`}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="设备ID">{device.id}</Descriptions.Item>
          <Descriptions.Item label="设备UUID">
            <Space>
              <Tag color="blue">{device.uuid}</Tag>
              <Button
                type="text"
                size="small"
                icon={<CopyOutlined />}
                onClick={handleCopyUuid}
                title="复制UUID"
              />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="设备令牌" span={2}>
            <Space>
              <Text code copyable={false} style={{ maxWidth: 300 }}>
                {formatToken(device.token)}
              </Text>
              <Button
                type="text"
                size="small"
                icon={tokenVisible ? <EyeOutlined /> : <EyeInvisibleOutlined />}
                onClick={toggleTokenVisibility}
                title={tokenVisible ? "隐藏令牌" : "显示令牌"}
              />
              <Button
                type="text"
                size="small"
                icon={<CopyOutlined />}
                onClick={handleCopyToken}
                title="复制令牌"
              />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="注册时间">
            {formatDate(device.registerTime)}
          </Descriptions.Item>
          <Descriptions.Item label="最后心跳时间">
            {formatDate(device.lastPingTime)}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Drawer
        title="编辑设备"
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <DeviceForm
          device={device}
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </div>
  );
};

export default DeviceDetail;