import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  message,
  Card,
  Input,
  Typography,
  Tooltip,
} from 'antd';
import {
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { deviceApi } from '../../services/api';
import type { Device, ListParams } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import { DEFAULT_PAGE_SIZE } from '../../utils/constants';

const { Search } = Input;
const { Text, Link } = Typography;

const DeviceList: React.FC = () => {
  const navigate = useNavigate();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  // 已移除令牌显示逻辑

  // 相对时间格式化：秒/分钟/小时/天前
  const formatRelativeTime = (dateStr: string | undefined): string => {
    if (!dateStr) return '-';
    const time = new Date(dateStr).getTime();
    if (isNaN(time)) return '-';
    const now = Date.now();
    let diffSec = Math.floor((now - time) / 1000);
    if (diffSec < 0) diffSec = 0;
    if (diffSec < 60) return `${diffSec}秒前`;
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}分钟前`;
    const diffHour = Math.floor(diffMin / 60);
    if (diffHour < 24) return `${diffHour}小时前`;
    const diffDay = Math.floor(diffHour / 24);
    return `${diffDay}天前`;
  };

  // 获取设备列表
  const fetchDevices = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await deviceApi.list(params);
      setDevices(response.devices || []);
      setTotal(response.total || 0);
    } catch (error) {
      handleApiError(error, '获取设备列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDevices();
  }, [current, pageSize]);

  // 处理删除设备
  const handleDelete = (device: Device) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除设备 "${device.uuid}"`),
      onOk: async () => {
        try {
          await deviceApi.delete(device.id);
          message.success('删除成功');
          fetchDevices();
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理查看详情
  const handleView = (device: Device) => {
    navigate(`/devices/${device.id}`);
  };

  // 表格列配置
  const columns: ColumnsType<Device> = [
    {
      title: '设备UUID',
      dataIndex: 'uuid',
      key: 'uuid',
      width: 200,
      ellipsis: true,
      render: (uuid: string, record: Device) => (
        <Link onClick={() => handleView(record)} ellipsis style={{ maxWidth: 150 }}>
          {uuid}
        </Link>
      ),
    },
    {
      title: '设备名称',
      dataIndex: 'name',
      key: 'name',
      width: 250,
      ellipsis: true,
      render: (name: string) => (
        <Text copyable={false} ellipsis style={{ maxWidth: 200 }}>
          {name}
        </Text>
      ),
    },
    {
      title: '注册时间',
      dataIndex: 'registerTime',
      key: 'registerTime',
      width: 180,
      render: (date: string) => formatDate(date),
    },
    {
      title: '最后心跳时间',
      dataIndex: 'lastPingTime',
      key: 'lastPingTime',
      width: 180,
      render: (date: string) => (
        <Tooltip title={`最后心跳时间：${formatDate(date)}`}>
          <span>{formatRelativeTime(date)}</span>
        </Tooltip>
      ),
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
            onClick={fetchDevices}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Search
            placeholder="搜索设备UUID"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchDevices}
          />
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={devices}
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

export default DeviceList;