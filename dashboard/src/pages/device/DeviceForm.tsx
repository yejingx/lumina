import React from 'react';
import { Form, Input, Button, Space, message } from 'antd';
import { deviceApi } from '../../services/api';
import type { DeviceSpec } from '../../types';
import { handleApiError } from '../../utils/helpers';

export interface DeviceFormProps {
  device?: DeviceSpec | null;
  onSubmit: () => void;
  onCancel: () => void;
}

const DeviceForm: React.FC<DeviceFormProps> = ({ device, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);
  const isEditing = !!device;

  // 初始化表单数据
  React.useEffect(() => {
    if (device) {
      form.setFieldsValue({
        uuid: device.uuid,
        token: device.token,
      });
    }
  }, [device, form]);

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      if (isEditing && device) {
        // 注意：根据API接口，设备可能不支持直接更新，这里只是展示结构
        // 实际实现需要根据后端API调整
        message.info('设备信息更新功能暂未实现');
        onSubmit();
      } else {
        // 创建设备通常通过注册接口
        message.info('设备创建功能暂未实现');
        onSubmit();
      }
    } catch (error) {
      handleApiError(error, isEditing ? '更新设备失败' : '创建设备失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
    >
      <Form.Item
        name="uuid"
        label="设备UUID"
        rules={[
          { required: true, message: '请输入设备UUID' },
          { max: 100, message: '设备UUID最多100个字符' },
        ]}
      >
        <Input placeholder="请输入设备UUID" disabled={isEditing} />
      </Form.Item>

      <Form.Item
        name="token"
        label="设备令牌"
        rules={[
          { required: true, message: '请输入设备令牌' },
        ]}
      >
        <Input.Password placeholder="请输入设备令牌" disabled={isEditing} />
      </Form.Item>

      {isEditing && (
        <>
          <Form.Item label="注册时间">
            <Input value={device?.registerTime} disabled />
          </Form.Item>

          <Form.Item label="最后心跳时间">
            <Input value={device?.lastPingTime} disabled />
          </Form.Item>
        </>
      )}

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

export default DeviceForm;