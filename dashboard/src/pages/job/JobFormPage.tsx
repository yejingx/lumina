import React, { useState, useEffect } from 'react';
import { Card, Button, Typography } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { jobApi } from '../../services/api';
import type { JobSpec } from '../../types';
import { handleApiError } from '../../utils/helpers';
import JobForm from './JobForm';

const { Title } = Typography;

const JobFormPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [job, setJob] = useState<JobSpec | null>(null);
  const [loading, setLoading] = useState(false);
  const isEdit = Boolean(id);

  // 获取任务详情（编辑模式）
  const fetchJob = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const jobData = await jobApi.get(parseInt(id));
      setJob(jobData);
    } catch (error) {
      handleApiError(error, '获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isEdit) {
      fetchJob();
    }
  }, [id, isEdit]);

  const handleSubmit = () => {
    navigate('/jobs');
  };

  const handleCancel = () => {
    navigate('/jobs');
  };

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={handleCancel}
            style={{ marginRight: 16 }}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            {isEdit ? '编辑任务' : '创建任务'}
          </Title>
        </div>
      </Card>

      <Card loading={loading}>
        <JobForm
          job={job}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
        />
      </Card>
    </div>
  );
};

export default JobFormPage;