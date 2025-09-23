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
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { userApi } from '../../services/api';
import type { User, ListParams } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import { DEFAULT_PAGE_SIZE } from '../../utils/constants';
import UserForm from './UserForm';

const { Search } = Input;

const UserList: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 获取用户列表
  const fetchUsers = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await userApi.list(params);
      setUsers(response.users || []);
      setTotal(response.total || 0);
    } catch (error) {
      handleApiError(error, '获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [current, pageSize]);

  // 处理创建用户
  const handleCreate = () => {
    setDrawerVisible(true);
  };

  // 处理删除用户
  const handleDelete = (user: User) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除用户 "${user.username}"`),
      onOk: async () => {
        try {
          await userApi.delete(user.id);
          message.success('删除成功');
          fetchUsers();
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchUsers();
  };

  // 表格列配置
  const columns: ColumnsType<User> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      ellipsis: true,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
      render: (email: string) => email || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => formatDate(date),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (date: string) => formatDate(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Space size="small">
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
            创建用户
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchUsers}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Search
            placeholder="搜索用户名"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchUsers}
          />
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={users}
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
        title="创建用户"
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <UserForm
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </Card>
  );
};

export default UserList;