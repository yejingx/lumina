import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  message,
  Card,
  Input,
  Drawer,
  Typography,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { accessTokenApi } from '../../services/api';
import type { AccessToken, ListParams } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig, copyToClipboard } from '../../utils/helpers';
import { DEFAULT_PAGE_SIZE } from '../../utils/constants';
import AccessTokenForm from './AccessTokenForm';

const { Search } = Input;
const { Text } = Typography;

const AccessTokenList: React.FC = () => {
  const navigate = useNavigate();
  const [tokens, setTokens] = useState<AccessToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 获取接入凭证列表
  const fetchTokens = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await accessTokenApi.list(params);
      setTokens(response.accessTokens || []);
      setTotal(response.total || 0);
    } catch (error) {
      handleApiError(error, '获取接入凭证列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTokens();
  }, [current, pageSize]);

  // 处理创建接入凭证
  const handleCreate = () => {
    setDrawerVisible(true);
  };

  // 处理删除接入凭证
  const handleDelete = (token: AccessToken) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除接入凭证 "${token.accessToken}"`),
      onOk: async () => {
        try {
          await accessTokenApi.delete(token.id);
          message.success('删除成功');
          fetchTokens();
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理查看详情
  const handleView = (token: AccessToken) => {
    navigate(`/access-tokens/${token.id}`);
  };

  // 处理复制令牌
  const handleCopyToken = (token: string) => {
    copyToClipboard(token);
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchTokens();
  };

  // 表格列配置
  const columns: ColumnsType<AccessToken> = [
    {
      title: '接入凭证',
      dataIndex: 'accessToken',
      key: 'accessToken',
      width: 200,
      render: (token: string) => (
        <Space>
          <Text code copyable={false} ellipsis style={{ maxWidth: 120 }}>
            {token}
          </Text>
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => handleCopyToken(token)}
            title="复制令牌"
          />
        </Space>
      ),
    },
    {
      title: '设备UUID',
      dataIndex: 'deviceUuid',
      key: 'deviceUuid',
      width: 200,
      ellipsis: true,
      render: (deviceUuid: string) => deviceUuid || '未绑定',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 180,
      render: (date: string) => formatDate(date),
    },
    {
      title: '过期时间',
      dataIndex: 'expireTime',
      key: 'expireTime',
      width: 180,
      render: (date: string) => formatDate(date) || '永不过期',
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
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            创建接入凭证
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchTokens}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Search
            placeholder="搜索接入凭证"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchTokens}
          />
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={tokens}
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

      <Drawer
        title="创建接入凭证"
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <AccessTokenForm
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </Card>
  );
};

export default AccessTokenList;