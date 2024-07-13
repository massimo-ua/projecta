import React, { useEffect, useState } from 'react';
import {
  BuildOutlined,
  DollarOutlined,
  FileTextOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import { Layout, Menu } from 'antd';
import { Outlet, useNavigate, useParams } from 'react-router-dom';
import HomeLayout from '../../Layout.jsx';

const { Sider, Content } = Layout;

export function ProjectDetails() {
  const navigate = useNavigate();
  const [navMenuItems, setNavMenuItems] = useState([{
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
        key: 'total',
        label: 'Total',
        icon: <FileTextOutlined />,
      },
      {
        key: 'expenses',
        label: 'Expenses',
        icon: <DollarOutlined />,
      },
    ],
  }]);

  const onClick = (e) => {
    const { key } = e;
    navigate(key);
  };

  return (
    <HomeLayout>
      <Layout>
        <Sider width="12vw">
          <Menu
            onClick={onClick}
            style={{ height: '80vh' }}
            defaultSelectedKeys={['expenses']}
            mode="inline"
            items={navMenuItems}
          />
        </Sider>
        <Content>
          <Outlet />
        </Content>
      </Layout>
    </HomeLayout>
  );
}
