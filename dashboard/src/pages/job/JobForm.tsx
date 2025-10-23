import React from 'react';
import { Form, Input, Select, Button, Card, message, InputNumber, Divider, Space } from 'antd';
import { jobApi, deviceApi, workflowApi } from '../../services/api';
import type { Job, CreateJobRequest, JobKind, DetectOptions, VideoSegmentOptions, Workflow } from '../../types';
import { handleApiError } from '../../utils/helpers';
import { DeleteOutlined } from '@ant-design/icons';

const { Option } = Select;

interface JobFormProps {
  job?: Job | null;
  onSubmit: () => void;
  onCancel: () => void;
}

const JobForm: React.FC<JobFormProps> = ({ job, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);
  const [devices, setDevices] = React.useState<any[]>([]);
  const [workflows, setWorkflows] = React.useState<Workflow[]>([]);
  const [selectedKind, setSelectedKind] = React.useState<JobKind>('detect');

  // 获取设备列表
  const fetchDevices = async () => {
    try {
      const response = await deviceApi.list({ start: 0, limit: 100 });
      setDevices(response.devices || []);
    } catch (error) {
      console.error('获取设备列表失败:', error);
    }
  };

  // 获取工作流列表
  const fetchWorkflows = async () => {
    try {
      const response = await workflowApi.list({ start: 0, limit: 100 });
      setWorkflows(response.workflows || []);
    } catch (error) {
      console.error('获取工作流列表失败:', error);
    }
  };

  React.useEffect(() => {
    fetchDevices();
    fetchWorkflows();
  }, []);

  // 初始化表单数据
  React.useEffect(() => {
    if (job) {
      const initialValues: any = {
        kind: job.kind,
        input: job.input,
        deviceId: job.device.id,
        workflowId: job.workflowId,
        query: job.query,
        resultFilter: job.resultFilter,
      };

      // Initialize detect options if present
      if (job.detect) {
        initialValues.detect = job.detect;
      }

      // Initialize video segment options if present
      if (job.videoSegment) {
        initialValues.videoSegment = job.videoSegment;
      }

      form.setFieldsValue(initialValues);
      setSelectedKind(job.kind);
    }
  }, [job, form]);

  // Handle kind change
  const handleKindChange = (value: JobKind) => {
    setSelectedKind(value);
    // Clear options when switching kind
    if (value === 'detect') {
      form.setFieldsValue({ videoSegment: undefined });
    } else if (value === 'video_segment') {
      form.setFieldsValue({ detect: undefined });
    }
  };

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      const data: CreateJobRequest = {
        kind: values.kind,
        input: values.input || '',
        workflowId: values.workflowId,
        query: values.query,
        deviceId: values.deviceId,
        resultFilter: values.resultFilter,
      };

      // Add specific options based on kind
      if (values.kind === 'detect' && values.detect) {
        data.detect = values.detect;
      } else if (values.kind === 'video_segment' && values.videoSegment) {
        data.videoSegment = values.videoSegment;
      }

      if (job) {
        await jobApi.update(job.id, data);
        message.success('更新任务成功');
      } else {
        await jobApi.create(data);
        message.success('创建任务成功');
      }

      onSubmit();
    } catch (error) {
      handleApiError(error, job ? '更新任务失败' : '创建任务失败');
    } finally {
      setLoading(false);
    }
  };

  // Render detect options form
  const renderDetectOptions = () => (
    <>
      <Divider>检测参数</Divider>
      <Form.Item
        name={['detect', 'modelName']}
        label="模型名称"
        rules={[{ required: true, message: '请输入模型名称' }]}
      >
        <Input placeholder="请输入模型名称" />
      </Form.Item>

      <Form.Item
        name={['detect', 'labels']}
        label="标签"
      >
        <Input placeholder="请输入标签（可选）" />
      </Form.Item>

      <Form.Item
        name={['detect', 'confThreshold']}
        label="置信度阈值"
      >
        <InputNumber
          min={0}
          max={1}
          step={0.01}
          placeholder="置信度阈值（0-1）"
          defaultValue={0.25}
          style={{ width: '100%' }}
        />
      </Form.Item>

      <Form.Item
        name={['detect', 'iouThreshold']}
        label="IoU阈值"
      >
        <InputNumber
          min={0}
          max={1}
          step={0.01}
          defaultValue={0.45}
          placeholder="IoU阈值（0-1）"
          style={{ width: '100%' }}
        />
      </Form.Item>

      <Form.Item
        name={['detect', 'interval']}
        label="检测间隔"
      >
        <InputNumber
          min={1}
          placeholder="检测间隔（ms）"
          defaultValue={1000}
          style={{ width: '100%' }}
        />
      </Form.Item>

      {/* 新增触发次数 */}
      <Form.Item
        name={['detect', 'triggerCount']}
        label="触发次数"
      >
        <InputNumber
          min={1}
          placeholder="触发次数"
          defaultValue={1}
          style={{ width: '100%' }}
        />
      </Form.Item>

      {/* 新增触发间隔（秒） */}
      <Form.Item
        name={['detect', 'triggerInterval']}
        label="触发间隔"
      >
        <InputNumber
          min={1}
          placeholder="触发间隔（秒）"
          defaultValue={30}
          style={{ width: '100%' }}
        />
      </Form.Item>
    </>
  );

  // Render video segment options form
  const renderVideoSegmentOptions = () => (
    <>
      <Divider>视频分割参数</Divider>
      <Form.Item
        name={['videoSegment', 'interval']}
        label="分割间隔"
      >
        <InputNumber
          min={1}
          placeholder="分割间隔（秒）"
          style={{ width: '100%' }}
        />
      </Form.Item>
    </>
  );

  // Render result filter form
  const renderResultFilter = () => (
    <>
      <Divider>结果过滤</Divider>
      <Form.Item name={['resultFilter', 'combineOp']} label="组合逻辑">
        <Select placeholder="请选择逻辑">
          <Option value="and">并且 (AND)</Option>
          <Option value="or">或者 (OR)</Option>
        </Select>
      </Form.Item>

      <Form.List name={['resultFilter', 'conditions']}>
        {(fields, { add, remove }) => (
          <div>
            {fields.map((field) => (
              <Card key={field.key} size="small" style={{ marginBottom: 8 }}>
                <Space size="small" align="center">
                  {/* <Form.Item
                    {...field}
                    name={[field.name, 'field']}
                    fieldKey={[field.fieldKey!, 'field']}
                    rules={[{ required: true, message: '请输入字段名' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input placeholder="字段名，例如 answer 或 label" style={{ width: 160 }} />
                  </Form.Item> */}

                  <Form.Item
                    {...field}
                    name={[field.name, 'op']}
                    fieldKey={[field.fieldKey!, 'op']}
                    rules={[{ required: true, message: '请选择操作符' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Select placeholder="选择操作符" style={{ width: 120 }}>
                      <Option value="eq">等于</Option>
                      <Option value="ne">不等于</Option>
                      <Option value="in">包含于</Option>
                      <Option value="not_in">不包含于</Option>
                      <Option value="contains">包含</Option>
                      <Option value="not_contains">不包含</Option>
                      <Option value="starts_with">开头为</Option>
                      <Option value="ends_with">结尾为</Option>
                      <Option value="empty">为空</Option>
                      <Option value="not_empty">不为空</Option>
                    </Select>
                  </Form.Item>

                  <Form.Item
                    {...field}
                    name={[field.name, 'value']}
                    fieldKey={[field.fieldKey!, 'value']}
                    style={{ marginBottom: 0 }}
                  >
                    <Input placeholder="匹配值（empty/not_empty 可留空）" style={{ width: 370 }} />
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
    </>
  );

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{
        kind: 'detect',
      }}
    >
      <Form.Item
        name="kind"
        label="任务类型"
        rules={[{ required: true, message: '请选择任务类型' }]}
      >
        <Select placeholder="请选择任务类型" onChange={handleKindChange}>
          <Option value="detect">检测任务</Option>
          <Option value="video_segment">视频分割</Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="input"
        label="输入"
        rules={[{ required: true, message: '请输入任务输入' }]}
      >
        <Input placeholder="请输入任务输入" />
      </Form.Item>

      <Form.Item
        name="deviceId"
        label="关联设备"
        rules={[{ required: true, message: '请选择关联设备' }]}
      >
        <Select placeholder="请选择设备" allowClear>
          {devices.map((device) => (
            <Option key={device.id} value={device.id}>
              {device.name} ({device.uuid})
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item
        name="workflowId"
        label="工作流"
      >
        <Select placeholder="请选择工作流（可选）" allowClear>
          {workflows.map((workflow) => (
            <Option key={workflow.id} value={workflow.id}>
              {workflow.name} ({workflow.uuid})
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item
        name="query"
        label="查询设置"
      >
        <Input.TextArea
          placeholder="请输入查询设置（可选）"
          rows={3}
          maxLength={1000}
          showCount
        />
      </Form.Item>

      {/* Render specific options based on selected kind */}
      {selectedKind === 'detect' && renderDetectOptions()}
      {selectedKind === 'video_segment' && renderVideoSegmentOptions()}

      {renderResultFilter()}

      <Divider />

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} style={{ marginRight: 8}}>
          {job ? '更新' : '创建'}
        </Button>
        <Button onClick={onCancel}>
          取消
        </Button>
      </Form.Item>
    </Form>
  );
};

export default JobForm;