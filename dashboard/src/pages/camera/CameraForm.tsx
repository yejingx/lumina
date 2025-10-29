import React, { useState, useEffect } from 'react';
import { Form, Input, Select, InputNumber, Button, Card, message, Space } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { cameraApi, deviceApi } from '../../services/api';
import type { CameraSpec, CreateCameraRequest, UpdateCameraRequest, DeviceSpec } from '../../types';

const { Option } = Select;

type CameraFormMode = 'page' | 'drawer';

interface CameraFormProps {
  mode?: CameraFormMode;
  cameraId?: number | string;
  onSubmit?: () => void;
  onCancel?: () => void;
}

const CameraForm: React.FC<CameraFormProps> = ({ mode = 'page', cameraId, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [camera, setCamera] = useState<CameraSpec | null>(null);
  const [devices, setDevices] = useState<DeviceSpec[]>([]);
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();

  const isDrawer = mode === 'drawer';
  const effectiveId = isDrawer ? cameraId : params.id;
  const isEdit = !!effectiveId && effectiveId !== 'new';

  const fetchCamera = async () => {
    if (!isEdit || !effectiveId) return;
    setLoading(true);
    try {
      const response = await cameraApi.get(parseInt(String(effectiveId)));
      setCamera(response);
      form.setFieldsValue({
        name: response.name,
        protocol: response.protocol,
        ip: response.ip,
        port: response.port,
        path: response.path,
        username: response.username,
        password: response.password,
        bindDeviceId: response.bindDevice?.id,
      });
    } catch (error) {
      message.error('获取摄像头信息失败');
      // eslint-disable-next-line no-console
      console.error('Error fetching camera:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCamera();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [effectiveId]);

  // 获取设备列表
  const fetchDevices = async () => {
    try {
      const resp = await deviceApi.list({ start: 0, limit: 100 });
      setDevices(resp.devices || []);
    } catch (error) {
      // eslint-disable-next-line no-console
      console.error('获取设备列表失败:', error);
    }
  };

  useEffect(() => {
    fetchDevices();
  }, []);

  const handleFinish = async (values: any) => {
    setLoading(true);
    try {
      if (isEdit && camera) {
        const updateData: UpdateCameraRequest = {
          name: values.name,
          protocol: values.protocol,
          ip: values.ip,
          port: values.port,
          path: values.path,
          username: values.username,
          password: values.password,
          bindDeviceId: values.bindDeviceId,
        };
        await cameraApi.update(camera.id, updateData);
        message.success('更新成功');
      } else {
        const createData: CreateCameraRequest = {
          name: values.name,
          protocol: values.protocol,
          ip: values.ip,
          port: values.port,
          path: values.path,
          username: values.username,
          password: values.password,
          bindDeviceId: values.bindDeviceId,
        };
        await cameraApi.create(createData);
        message.success('创建成功');
      }

      if (isDrawer) {
        onSubmit && onSubmit();
      } else {
        navigate('/camera');
      }
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
      // eslint-disable-next-line no-console
      console.error('Error saving camera:', error);
    } finally {
      setLoading(false);
    }
  };

  const formNode = (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleFinish}
      initialValues={{
        protocol: 'rtsp',
        port: 554,
      }}
    >
      <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入摄像头名称' }]}>
        <Input placeholder="请输入摄像头名称" />
      </Form.Item>

      <Form.Item label="协议" name="protocol" rules={[{ required: true, message: '请选择协议' }]}>
        <Select placeholder="请选择协议">
          <Option value="rtsp">RTSP</Option>
          <Option value="rtmp">RTMP</Option>
        </Select>
      </Form.Item>

      <Form.Item
        label="IP地址"
        name="ip"
        rules={[
          { required: true, message: '请输入IP地址' },
          {
            pattern:
              /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
            message: '请输入有效的IP地址',
          },
        ]}
      >
        <Input placeholder="请输入IP地址，如：192.168.1.100" />
      </Form.Item>

      <Form.Item
        label="端口"
        name="port"
        rules={[
          { required: true, message: '请输入端口号' },
          { type: 'number', min: 1, max: 65535, message: '端口号必须在1-65535之间' },
        ]}
      >
        <InputNumber placeholder="请输入端口号" style={{ width: '100%' }} min={1} max={65535} />
      </Form.Item>

      <Form.Item label="路径" name="path">
        <Input placeholder="请输入路径，如：/stream1" />
      </Form.Item>

      <Form.Item label="用户名" name="username">
        <Input placeholder="请输入用户名（可选）" />
      </Form.Item>

      <Form.Item label="密码" name="password">
        <Input.Password placeholder="请输入密码（可选）" />
      </Form.Item>

      <Form.Item label="绑定设备" name="bindDeviceId">
        <Select placeholder="请选择设备（可选）" allowClear>
          {devices.map((device) => (
            <Option key={device.id} value={device.id}>
              {device.name} ({device.uuid})
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={loading} icon={<SaveOutlined />}>
            {isEdit ? '更新' : '创建'}
          </Button>
          {isDrawer ? (
            <Button onClick={() => onCancel && onCancel()}>取消</Button>
          ) : (
            <Button onClick={() => navigate('/cameras')}>取消</Button>
          )}
        </Space>
      </Form.Item>
    </Form>
  );

  if (isDrawer) {
    return formNode;
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/camera')}>
          返回列表
        </Button>
      </div>

      <Card title={isEdit ? '编辑摄像头' : '添加摄像头'} loading={loading}>
        {formNode}
      </Card>
    </div>
  );
};

export default CameraForm;