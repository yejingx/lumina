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
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { workflowApi } from '../../services/api';
import type { WorkflowSpec } from '../../types';
import { formatDate, handleApiError, getDeleteConfirmConfig } from '../../utils/helpers';
import WorkflowForm from './WorkflowForm';

// 操作符中文标签映射
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

// 结果过滤内联文本生成函数
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

const WorkflowDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [workflow, setWorkflow] = useState<WorkflowSpec | null>(null);
  const [loading, setLoading] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 获取工作流详情
  const fetchWorkflowDetail = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const workflowId = parseInt(id, 10);
      const response = await workflowApi.get(workflowId);
      setWorkflow(response);
    } catch (error) {
      handleApiError(error, '获取工作流详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkflowDetail();
  }, [id]);

  // 处理编辑
  const handleEdit = () => {
    setDrawerVisible(true);
  };

  // 处理删除
  const handleDelete = () => {
    if (!workflow) return;
    
    Modal.confirm({
      ...getDeleteConfirmConfig(`删除工作流 "${workflow.name}"`),
      onOk: async () => {
        try {
          await workflowApi.delete(workflow.id);
          message.success('删除成功');
          navigate('/workflows');
        } catch (error) {
          handleApiError(error, '删除失败');
        }
      },
    });
  };

  // 处理表单提交
  const handleFormSubmit = () => {
    setDrawerVisible(false);
    fetchWorkflowDetail();
  };

  // 返回列表
  const handleBack = () => {
    navigate('/workflows');
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!workflow) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <p>工作流不存在或已被删除</p>
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
              onClick={fetchWorkflowDetail}
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

      <Card title={`工作流详情 - ${workflow.name}`}>
        <Descriptions column={2} bordered>
          <Descriptions.Item label="工作流名称">{workflow.name}</Descriptions.Item>
          <Descriptions.Item label="UUID" span={2}>
            <Tag color="blue">{workflow.uuid || '-'}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="API密钥" span={2}>
            <Tag color="orange">{workflow.key ? '已设置' : '未设置'}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="API地址" span={2}>
            <a href={workflow.endpoint} target="_blank" rel="noopener noreferrer">
              {workflow.endpoint || '-'}
            </a>
          </Descriptions.Item>
          <Descriptions.Item label="超时时间">
            {workflow.timeout ? `${workflow.timeout} 毫秒` : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {formatDate(workflow.createTime)}
          </Descriptions.Item>
          <Descriptions.Item label="查询设置" span={2}>
            {workflow.query ? (
              <div style={{
                maxHeight: '100px',
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                backgroundColor: '#f5f5f5',
                padding: '8px',
                borderRadius: '4px'
              }}>
                {workflow.query}
              </div>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="结果过滤" span={2}>
            {getResultFilterText(workflow?.resultFilter)}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Drawer
        title="编辑工作流"
        width={600}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        destroyOnClose
      >
        <WorkflowForm
          workflow={workflow}
          onSubmit={handleFormSubmit}
          onCancel={() => setDrawerVisible(false)}
        />
      </Drawer>
    </div>
  );
};

export default WorkflowDetail;