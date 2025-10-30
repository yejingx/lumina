import React, { useState, useEffect } from 'react';
import { Table, Button, Space, message, Modal, Card, Tag, Drawer } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, ReloadOutlined, VideoCameraOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { cameraApi } from '../../services/api';
import { formatDate } from '../../utils/helpers';
import CameraForm from './CameraForm';
import type { CameraSpec, ListParams } from '../../types';

const CameraList: React.FC = () => {
  const [cameras, setCameras] = useState<CameraSpec[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ start: 0, limit: 10 });
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingCameraId, setEditingCameraId] = useState<number | null>(null);
  const navigate = useNavigate();

  const fetchCameras = async (params: ListParams = pagination) => {
    setLoading(true);
    try {
      const response = await cameraApi.list(params);
      setCameras(response.items);
      setTotal(response.total);
    } catch (error) {
      message.error('获取摄像头列表失败');
      console.error('Error fetching cameras:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCameras();
  }, []);

  const handleDelete = (camera: CameraSpec) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除摄像头 "${camera.name}" 吗？`,
      onOk: async () => {
        try {
          await cameraApi.delete(camera.id);
          message.success('删除成功');
          fetchCameras();
        } catch (error) {
          message.error('删除失败');
          console.error('Error deleting camera:', error);
        }
      },
    });
  };

  const getProtocolColor = (protocol: string) => {
    switch (protocol.toLowerCase()) {
      case 'rtsp':
        return 'blue';
      case 'rtmp':
        return 'volcano';
      default:
        return 'default';
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: CameraSpec) => (
        <Button
          type="link"
          onClick={() => navigate(`/cameras/${record.id}`)}
          style={{ padding: 0 }}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (protocol: string) => (
        <Tag color={getProtocolColor(protocol)}>{protocol.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: '端口',
      dataIndex: 'port',
      key: 'port',
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      render: (path: string) => path || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      render: (time: string) => formatDate(time),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: CameraSpec) => (
        <Space size="small">
          <Button
            type="text"
            icon={<VideoCameraOutlined />}
            onClick={() => navigate(`/cameras/${record.id}?tab=preview`)}
          >
            预览
          </Button>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/cameras/${record.id}`)}
          >
            查看
          </Button>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => { setEditingCameraId(record.id); setDrawerVisible(true); }}
          >
            编辑
          </Button>
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
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
            onClick={() => { setEditingCameraId(null); setDrawerVisible(true); }}
          >
            添加摄像头
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => fetchCameras()}
          >
            刷新
          </Button>
        </Space>
      </div>
      <Table
        columns={columns}
        dataSource={cameras}
        rowKey="id"
        loading={loading}
        pagination={{
          current: Math.floor(pagination.start / pagination.limit) + 1,
          pageSize: pagination.limit,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) =>
            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          onChange: (page, pageSize) => {
            const newPagination = {
              start: (page - 1) * pageSize,
              limit: pageSize,
            };
            setPagination(newPagination);
            fetchCameras(newPagination);
          },
        }}
      />

      <Drawer
        title={editingCameraId ? '编辑摄像头' : '添加摄像头'}
        width={600}
        open={drawerVisible}
        onClose={() => { setDrawerVisible(false); setEditingCameraId(null); }}
        destroyOnClose
      >
        <CameraForm
          mode="drawer"
          cameraId={editingCameraId ?? undefined}
          onSubmit={() => {
            setDrawerVisible(false);
            setEditingCameraId(null);
            fetchCameras();
          }}
          onCancel={() => { setDrawerVisible(false); setEditingCameraId(null); }}
        />
      </Drawer>
    </Card>
  );
};

export default CameraList;