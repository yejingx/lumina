import React from 'react';
import { Form, Input, Button, Space, message } from 'antd';
import { userApi } from '../../services/api';
import type { CreateUserRequest } from '../../types';
import { handleApiError } from '../../utils/helpers';

interface UserFormProps {
  onSubmit: () => void;
  onCancel: () => void;
}

const UserForm: React.FC<UserFormProps> = ({ onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      const data: CreateUserRequest = {
        username: values.username,
        email: values.email,
        password: values.password,
      };

      await userApi.create(data);
      message.success('创建用户成功');
      onSubmit();
    } catch (error) {
      handleApiError(error, '创建用户失败');
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
        name="username"
        label="用户名"
        rules={[
          { required: true, message: '请输入用户名' },
          { min: 3, message: '用户名至少3个字符' },
          { max: 50, message: '用户名最多50个字符' },
        ]}
      >
        <Input placeholder="请输入用户名" />
      </Form.Item>

      <Form.Item
        name="email"
        label="邮箱"
        rules={[
          { required: false },
          { type: 'email', message: '请输入有效的邮箱地址' },
        ]}
      >
        <Input placeholder="请输入邮箱地址" />
      </Form.Item>

      <Form.Item
        name="password"
        label="密码"
        rules={[
          { required: true, message: '请输入密码' },
          { min: 6, message: '密码至少6个字符' },
        ]}
      >
        <Input.Password placeholder="请输入密码" />
      </Form.Item>

      <Form.Item
        name="confirmPassword"
        label="确认密码"
        dependencies={['password']}
        rules={[
          { required: true, message: '请确认密码' },
          ({ getFieldValue }) => ({
            validator(_, value) {
              if (!value || getFieldValue('password') === value) {
                return Promise.resolve();
              }
              return Promise.reject(new Error('两次输入的密码不一致'));
            },
          }),
        ]}
      >
        <Input.Password placeholder="请再次输入密码" />
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

export default UserForm;