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

// 操作符中文标签映射（与详情页一致）
const OPERATOR_LABELS: Record<string, string> = {
  eq: '等于',
  ne: '不等于',
  in: '包含于',
  not_in: '不包含于',
  contains: '包含',
  not_contains: '不包含',
  starts_with: '开头为',
  ends_with: '结尾为',
  empty: '为空',
  not_empty: '不为空',
};

// 结果过滤内联文本
const getResultFilterText = (rf?: any) => {
  const rfLocal = rf;
  if (!rfLocal) return '-';
  const combineText = rfLocal.combineOp === 'and'
    ? '并且 (AND)'
    : rfLocal.combineOp === 'or'
      ? '或者 (OR)'
      : (rfLocal.combineOp || '-');
  const conditions = rfLocal.conditions || [];
  if (conditions.length === 0) return combineText;
  const condText = conditions.map((c: any) => {
    const opLabel = OPERATOR_LABELS[c.op] || c.op;
    const field = c.field || '查询结果';
    const value = (c.value ?? '-');
    return `${field} ${opLabel} ${value}`;
  }).join('；');
  return `${combineText}；${condText}`;
};

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
      render: (name: string, record: Workflow) => (
        <Button type="link" style={{ padding: 0 }} onClick={() => handleView(record)}>
          {name}
        </Button>
      ),
    },
    {
      title: '查询问题',
      dataIndex: 'query',
      key: 'uuid',
      ellipsis: false,
      render: (text: string) => (
        text ? (
          <pre
            style={{
              maxHeight: 120,
              overflow: 'auto',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              backgroundColor: '#f5f5f5',
              padding: 8,
              borderRadius: 4,
              fontFamily: 'monospace',
              margin: 0,
            }}
          >
            {text}
          </pre>
        ) : (
          '-'
        )
      ),
    },
    {
      title: '结果过滤',
      dataIndex: 'resultFilter',
      key: 'resultFilter',
      ellipsis: true,
      render: (_: any, record) => getResultFilterText(record.resultFilter),
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