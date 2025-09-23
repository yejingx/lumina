import React from 'react';
import { Form, Input, Button, Space, message, InputNumber } from 'antd';
import { workflowApi } from '../../services/api';
import type { Workflow, CreateWorkflowRequest, UpdateWorkflowRequest } from '../../types';
import { handleApiError } from '../../utils/helpers';

interface WorkflowFormProps {
  workflow?: Workflow | null;
  onSubmit: () => void;
  onCancel: () => void;
}

const WorkflowForm: React.FC<WorkflowFormProps> = ({ workflow, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);
  const isEditing = !!workflow;

  // 初始化表单数据
  React.useEffect(() => {
    if (workflow) {
      form.setFieldsValue({
        name: workflow.name,
        endpoint: workflow.endpoint || '',
        key: workflow.key || '',
        timeout: workflow.timeout || 3000,
      });
    }
  }, [workflow, form]);

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      if (isEditing && workflow) {
        const data: UpdateWorkflowRequest = {
          name: values.name,
          endpoint: values.endpoint,
          key: values.key,
          timeout: values.timeout,
        };
        await workflowApi.update(workflow.id, data);
        message.success('更新工作流成功');
      } else {
        const data: CreateWorkflowRequest = {
          name: values.name,
          endpoint: values.endpoint,
          key: values.key,
          timeout: values.timeout || 3000,
        };
        await workflowApi.create(data);
        message.success('创建工作流成功');
      }
      onSubmit();
    } catch (error) {
      handleApiError(error, isEditing ? '更新工作流失败' : '创建工作流失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{
        timeout: 3000,
      }}
    >
      <Form.Item
        name="name"
        label="工作流名称"
        rules={[
          { required: true, message: '请输入工作流名称' },
          { max: 100, message: '工作流名称最多100个字符' },
        ]}
      >
        <Input placeholder="请输入工作流名称" />
      </Form.Item>

      <Form.Item
        name="endpoint"
        label="端点地址"
        rules={[
          { required: true, message: '请输入端点地址' },
          { type: 'url', message: '请输入有效的URL地址' },
        ]}
      >
        <Input placeholder="请输入端点地址，如：https://api.example.com" />
      </Form.Item>

      <Form.Item
        name="key"
        label="API密钥"
        rules={[
          { required: true, message: '请输入API密钥' },
        ]}
      >
        <Input.Password placeholder="请输入API密钥" />
      </Form.Item>

      <Form.Item
        name="timeout"
        label="超时时间（毫秒）"
        rules={[
          { required: true, message: '请输入超时时间' },
          { type: 'number', min: 1000, max: 300000, message: '超时时间应在1000-300000毫秒之间' },
        ]}
      >
        <InputNumber
          placeholder="请输入超时时间"
          min={1000}
          max={300000}
          step={1000}
          style={{ width: '100%' }}
          addonAfter="毫秒"
        />
      </Form.Item>

      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={loading}>
            {isEditing ? '更新' : '创建'}
          </Button>
          <Button onClick={onCancel}>
            取消
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default WorkflowForm;