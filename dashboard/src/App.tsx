import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import MainLayout from './components/layout/MainLayout';
import { JobList, JobDetail } from './pages/job';
import { MessageList, MessageDetail } from './pages/message';
import { DeviceList, DeviceDetail } from './pages/device';
import { WorkflowList, WorkflowDetail } from './pages/workflow';
import { AccessTokenList, AccessTokenDetail } from './pages/access-token';
import { UserList } from './pages/user';
import JobFormPage from './pages/job/JobFormPage';

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Router>
        <Routes>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<Navigate to="/jobs" replace />} />

            {/* 任务路由 */}
            <Route path="jobs" element={<JobList />} />
            <Route path="jobs/new" element={<JobFormPage />} />
            <Route path="jobs/:id" element={<JobDetail />} />
            <Route path="jobs/:id/edit" element={<JobFormPage />} />

            {/* 消息路由 */}
            <Route path="messages" element={<MessageList />} />
            <Route path="messages/:id" element={<MessageDetail />} />

            {/* 设备路由 */}
            <Route path="devices" element={<DeviceList />} />
            <Route path="devices/:id" element={<DeviceDetail />} />

            {/* 工作流路由 */}
            <Route path="workflows" element={<WorkflowList />} />
            <Route path="workflows/:id" element={<WorkflowDetail />} />

            {/* 接入凭证路由 */}
            <Route path="access-tokens" element={<AccessTokenList />} />
            <Route path="access-tokens/:id" element={<AccessTokenDetail />} />

            {/* 用户路由 */}
            <Route path="users" element={<UserList />} />
          </Route>
        </Routes>
      </Router>
    </ConfigProvider>
  );
}

export default App;
