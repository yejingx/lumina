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
  Modal,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import { accessTokenApi } from '../../services/api';
import type { AccessTokenSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig, copyToClipboard } from '../../utils/helpers';

const { Text } = Typography;

const AccessTokenDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [accessToken, setAccessToken] = useState<AccessTokenSpec | null>(null);
  const [loading, setLoading] = useState(false);

  // 获取接入凭证详情
  const fetchAccessTokenDetail = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const tokenId = parseInt(id, 10);
      const response = await accessTokenApi.get(tokenId);
      setAccessToken(response);
    } catch (error) {
      handleApiError(error, '获取接入凭证详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAccessTokenDetail();
  }, [id]);

  // 处理删除
  const handleDelete = () => {
    if (!accessToken) return;
    
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除接入凭证 "${accessToken.accessToken}"`),
      onOk: async () => {
        try {
          await accessTokenApi.delete(accessToken.id);
          message.success('删除成功');
          navigate('/access-tokens');
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理复制令牌
  const handleCopyToken = () => {
    if (accessToken) {
      copyToClipboard(accessToken.accessToken);
    }
  };

  // 返回列表
  const handleBack = () => {
    navigate('/access-tokens');
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!accessToken) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <p>接入凭证不存在或已被删除</p>
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
              onClick={fetchAccessTokenDetail}
            >
              刷新
            </Button>
          </Space>
          <Space>
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

      <Card title={`接入凭证详情 - ${accessToken.accessToken.substring(0, 8)}...`}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="接入凭证" span={2}>
            <Space>
              <Text code copyable={false} style={{ maxWidth: 300 }}>
                {accessToken.accessToken}
              </Text>
              <Button
                type="text"
                size="small"
                icon={<CopyOutlined />}
                onClick={handleCopyToken}
                title="复制令牌"
              />
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="设备UUID" span={2}>
            <Tag color={accessToken.deviceUuid ? "blue" : "default"}>
              {accessToken.deviceUuid || '未绑定'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {formatDate(accessToken.createTime)}
          </Descriptions.Item>
          <Descriptions.Item label="过期时间">
            {formatDate(accessToken.expireTime) || '永不过期'}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
};

export default AccessTokenDetail;