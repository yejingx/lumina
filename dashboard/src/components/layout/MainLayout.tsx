import React, { useState } from 'react';
import { Layout, Menu, theme } from 'antd';
import {
  UserOutlined,
  DesktopOutlined,
  FileTextOutlined,
  MessageOutlined,
  KeyOutlined,
  PartitionOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
  CameraOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [openKeysState, setOpenKeysState] = useState<string[]>([]);
  const navigate = useNavigate();
  const location = useLocation();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // 菜单项配置
  const menuItems = [
    {
      key: '/workflows',
      icon: <PartitionOutlined />,
      label: '工作流管理',
    },
    {
      key: '/jobs',
      icon: <FileTextOutlined />,
      label: '任务管理',
    },
    {
      key: '/messages',
      icon: <MessageOutlined />,
      label: '消息管理',
    },
    {
      key: 'devices-root',
      icon: <DesktopOutlined />,
      label: '设备管理',
      children: [
        {
          key: '/devices',
          icon: <DesktopOutlined />,
          label: '主机管理',
        },
        {
          key: '/access-tokens',
          icon: <KeyOutlined />,
          label: '接入凭证',
        },
        {
          key: '/cameras',
          icon: <CameraOutlined />,
          label: '摄像头管理',
        },
      ],
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: '/agent',
      icon: <RobotOutlined />,
      label: '智能助手',
    },
  ];

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  // 根据当前路径匹配选中与展开菜单
  const getMenuState = () => {
    const path = location.pathname;

    const dfs = (items: any[], parentKey?: string): { selected?: string; openKeys?: string[] } => {
      for (const item of items) {
        if (item.children && Array.isArray(item.children)) {
          const res = dfs(item.children, item.key);
          if (res.selected) {
            return { selected: res.selected, openKeys: parentKey ? [parentKey, item.key] : [item.key] };
          }
        }
        if (typeof item.key === 'string' && item.key.startsWith('/') && path.startsWith(item.key)) {
          return { selected: item.key, openKeys: parentKey ? [parentKey] : [] };
        }
      }
      return {};
    };

    const { selected, openKeys } = dfs(menuItems);
    const selectedKeys = selected ? [selected] : [];
    const routeOpenKeys = openKeys || [];
    const mergedOpenKeys = Array.from(new Set([...(openKeysState || []), ...routeOpenKeys]));
    return { selectedKeys, openKeys: mergedOpenKeys };
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div style={{
          height: 32,
          margin: 16,
          background: 'rgba(255, 255, 255, 0.2)',
          borderRadius: 6,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontWeight: 'bold',
          fontSize: collapsed ? 12 : 16,
        }}>
          {collapsed ? 'L' : 'Lumina'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getMenuState().selectedKeys}
          openKeys={getMenuState().openKeys}
          items={menuItems}
          onClick={handleMenuClick}
          onOpenChange={(keys) => setOpenKeysState(keys as string[])}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'margin-left 0.2s' }}>
        <Header
          style={{
            padding: 0,
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <div
            style={{
              fontSize: '16px',
              padding: '0 24px',
              cursor: 'pointer',
              transition: 'color 0.3s',
            }}
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
        </Header>
        <Content
          style={{
            margin: '24px 16px',
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
            overflow: 'auto',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;