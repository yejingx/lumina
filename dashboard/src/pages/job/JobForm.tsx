import React from 'react';
import { Form, Input, Select, Button, message, InputNumber, Divider } from 'antd';
import { jobApi, deviceApi, workflowApi, cameraApi } from '../../services/api';
import type { Job, CreateJobRequest, JobKind, DetectOptions, VideoSegmentOptions, Workflow, Camera } from '../../types';
import { handleApiError } from '../../utils/helpers';

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
  const [cameras, setCameras] = React.useState<Camera[]>([]);
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
    cameraApi.list({ start: 0, limit: 100 }).then((res) => setCameras(res.items || [])).catch(() => {});
  }, []);

  // 初始化表单数据
  React.useEffect(() => {
    if (job) {
      const initialValues: any = {
        kind: job.kind,
        cameraId: (job as any)?.camera?.id,
        deviceId: job.device.id,
        // 后端返回的字段由 workflowId 改为 workflow 对象，这里取其 id 作为初始值
        workflowId: (job as any)?.workflow?.id,
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
        cameraId: values.cameraId,
        workflowId: values.workflowId,
        deviceId: values.deviceId,
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
        name="cameraId"
        label="摄像头"
        rules={[{ required: true, message: '请选择摄像头' }]}
      >
        <Select placeholder="请选择摄像头" allowClear>
          {cameras.map((cam) => (
            <Option key={cam.id} value={cam.id}>
              {cam.name} ({cam.uuid})
            </Option>
          ))}
        </Select>
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


      {/* Render specific options based on selected kind */}
      {selectedKind === 'detect' && renderDetectOptions()}
      {selectedKind === 'video_segment' && renderVideoSegmentOptions()}

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