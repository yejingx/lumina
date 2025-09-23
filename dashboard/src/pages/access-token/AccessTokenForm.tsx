import React from 'react';
import { Form, Input, Button, Space, message, DatePicker } from 'antd';
import dayjs from 'dayjs';
import { accessTokenApi } from '../../services/api';
import type { CreateAccessTokenRequest } from '../../types';
import { handleApiError } from '../../utils/helpers';

interface AccessTokenFormProps {
  onSubmit: () => void;
  onCancel: () => void;
}

const AccessTokenForm: React.FC<AccessTokenFormProps> = ({ onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      const data: CreateAccessTokenRequest = {
        expireTime: values.expireTime ? values.expireTime.toISOString() : '',
      };

      await accessTokenApi.create(data);
      message.success('创建接入凭证成功');
      onSubmit();
    } catch (error) {
      handleApiError(error, '创建接入凭证失败');
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
        name="expireTime"
        label="过期时间"
        rules={[
          { required: true, message: '请选择过期时间' },
        ]}
      >
        <DatePicker
          showTime
          placeholder="选择过期时间"
          style={{ width: '100%' }}
          disabledDate={(current) => current && current < dayjs().endOf('day')}
        />
      </Form.Item>

      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={loading}>
            创建
          </Button>
          <Button onClick={onCancel}>
            取消
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default AccessTokenForm;