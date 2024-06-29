import React from 'react';
import { HomeLayout } from '../../Layout.jsx';
import { BuildOutlined, PieChartOutlined, TransactionOutlined } from '@ant-design/icons';
import { Layout, Menu } from 'antd';
import { Outlet, useNavigate, useParams } from 'react-router-dom';

const { Sider, Content } = Layout;

const navigation = [{
  key: 'resources',
  label: 'Resources',
  type: 'group',
  children: [
    {
      key: 'categories',
      label: 'Categories',
      icon: <PieChartOutlined />,
    },
    {
      key: 'types',
      label: 'Types',
      icon: <BuildOutlined />,
    },
  ],
}, {
  key: 'operations',
  label: 'Operations',
  type: 'group',
  children: [
    {
      key: 'expenses',
      label: 'Expenses',
      icon: <TransactionOutlined />,
    },
  ],
}];

export function ProjectDetails() {
  const navigate = useNavigate();
  const { projectId} = useParams();

  const onClick = (e) => {
    const { key } = e;
    navigate(key);
  };

  return (
    <HomeLayout>
      <Layout>
        <Sider width="300">
          <Menu
            onClick={onClick}
            style={{ height: '80vh' }}
            defaultSelectedKeys={['expenses']}
            mode="inline"
            items={navigation}
          />
        </Sider>
        <Content>
          <Outlet />
        </Content>
      </Layout>
    </HomeLayout>
  );
}
