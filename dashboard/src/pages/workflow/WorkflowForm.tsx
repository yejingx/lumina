import React from 'react';
import { Form, Input, Button, Space, message, InputNumber, Select, Divider, Card } from 'antd';
import { workflowApi } from '../../services/api';
import type { Workflow, CreateWorkflowRequest, UpdateWorkflowRequest } from '../../types';
import { handleApiError } from '../../utils/helpers';
import { DeleteOutlined } from '@ant-design/icons';

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
        modelName: workflow.modelName || '',
        timeout: workflow.timeout || 30000,
        query: workflow.query || '',
        resultFilter: workflow.resultFilter,
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
          query: values.query,
          resultFilter: values.resultFilter,
          modelName: values.modelName,
        };
        await workflowApi.update(workflow.id, data);
        message.success('更新工作流成功');
      } else {
        const data: CreateWorkflowRequest = {
          name: values.name,
          endpoint: values.endpoint,
          key: values.key,
          timeout: values.timeout || 30000,
          query: values.query,
          resultFilter: values.resultFilter,
          modelName: values.modelName,
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
        timeout: 30000,
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
        label="API地址"
        rules={[
          { required: true, message: '请输入API地址' },
          { type: 'url', message: '请输入有效的URL地址' },
        ]}
      >
        <Input placeholder="请输入API地址，如：https://api.example.com" />
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
        name="modelName"
        label="模型名称"
        rules={[
          { required: true, message: '请输入模型名称' },
        ]}
      >
        <Input placeholder="请输入模型名称" />
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

      <Form.Item
        name="query"
        label="查询设置"
        rules={[{ required: true, message: '请输入查询设置' }]}
      >
        <Input.TextArea
          placeholder="请输入查询设置（必填）"
          rows={3}
          maxLength={1000}
          showCount
        />
      </Form.Item>

      <Divider>结果过滤</Divider>
      <Form.Item name={["resultFilter", "combineOp"]} label="组合逻辑">
        <Select placeholder="请选择逻辑">
          <Select.Option value="and">并且 (AND)</Select.Option>
          <Select.Option value="or">或者 (OR)</Select.Option>
        </Select>
      </Form.Item>

      <Form.List name={["resultFilter", "conditions"]}>
        {(fields, { add, remove }) => (
          <div>
            {fields.map((field) => (
              <Card key={field.key} size="small" style={{ marginBottom: 8 }}>
                <Space align="center" size="small">
                  <Form.Item
                    {...field}
                    name={[field.name, 'op']}
                    fieldKey={[field.fieldKey!, 'op']}
                    rules={[{ required: true, message: '请选择操作符' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Select placeholder="选择操作符" style={{ width: 120 }}>
                      <Select.Option value="eq">等于</Select.Option>
                      <Select.Option value="ne">不等于</Select.Option>
                      <Select.Option value="in">包含于</Select.Option>
                      <Select.Option value="not_in">不包含于</Select.Option>
                      <Select.Option value="contains">包含</Select.Option>
                      <Select.Option value="not_contains">不包含</Select.Option>
                      <Select.Option value="starts_with">开头为</Select.Option>
                      <Select.Option value="ends_with">结尾为</Select.Option>
                      <Select.Option value="empty">为空</Select.Option>
                      <Select.Option value="not_empty">不为空</Select.Option>
                    </Select>
                  </Form.Item>

                  <Form.Item
                    {...field}
                    name={[field.name, 'value']}
                    fieldKey={[field.fieldKey!, 'value']}
                    style={{ marginBottom: 0 }}
                  >
                    <Input placeholder="匹配值（empty/not_empty 可留空）" style={{ width: 360 }} />
                  </Form.Item>

                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => remove(field.name)}
                    title="删除条件"
                  >
                  </Button>
                </Space>
              </Card>
            ))}
            <Button type="dashed" onClick={() => add()} block>
              新增条件
            </Button>
          </div>
        )}
      </Form.List>

      <Divider />

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