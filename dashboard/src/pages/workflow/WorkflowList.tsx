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
  EyeOutlined,
  EditOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { workflowApi } from '../../services/api';
import type { Workflow, ListParams } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import { DEFAULT_PAGE_SIZE } from '../../utils/constants';
import WorkflowForm from './WorkflowForm';

const { Search } = Input;

const WorkflowList: React.FC = () => {
  const navigate = useNavigate();
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchText, setSearchText] = useState('');
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingWorkflow, setEditingWorkflow] = useState<Workflow | null>(null);

  // 获取工作流列表
  const fetchWorkflows = async () => {
    setLoading(true);
    try {
      const params: ListParams = {
        start: (current - 1) * pageSize,
        limit: pageSize,
      };
      const response = await workflowApi.list(params);
      setWorkflows(response.workflows || []);
      setTotal(response.total || 0);
    } catch (error) {
      handleApiError(error, '获取工作流列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkflows();
  }, [current, pageSize]);

  // 处理创建工作流
  const handleCreate = () => {
    setEditingWorkflow(null);
    setDrawerVisible(true);
  };

  // 处理编辑工作流
  const handleEdit = (workflow: Workflow) => {
    setEditingWorkflow(workflow);
    setDrawerVisible(true);
  };

  // 处理删除工作流
  const handleDelete = (workflow: Workflow) => {
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除工作流 "${workflow.name}"`),
      onOk: async () => {
        try {
          await workflowApi.delete(workflow.id);
          message.success('删除成功');
          fetchWorkflows();
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理查看详情
  const handleView = (workflow: Workflow) => {
    navigate(`/workflows/${workflow.id}`);
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    setEditingWorkflow(null);
    fetchWorkflows();
  };

  // 表格列配置
  const columns: ColumnsType<Workflow> = [
    {
      title: '工作流名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'UUID',
      dataIndex: 'uuid',
      key: 'uuid',
      ellipsis: true,
      render: (text: string) => text || '-',
    },
    {
      title: '端点地址',
      dataIndex: 'endpoint',
      key: 'endpoint',
      ellipsis: true,
      render: (text: string) => text || '-',
    },
    {
      title: '超时时间(ms)',
      dataIndex: 'timeout',
      key: 'timeout',
      width: 120,
      render: (timeout: number) => timeout || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 180,
      render: (date: string) => formatDate(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
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
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            title="编辑"
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
            创建工作流
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchWorkflows}
          >
            刷新
          </Button>
        </Space>
        <Space style={{ float: 'right' }}>
          <Search
            placeholder="搜索工作流名称"
            allowClear
            style={{ width: 200 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onSearch={fetchWorkflows}
          />
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={workflows}
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
        title={editingWorkflow ? '编辑工作流' : '创建工作流'}
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <WorkflowForm
          workflow={editingWorkflow}
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </Card>
  );
};

export default WorkflowList;