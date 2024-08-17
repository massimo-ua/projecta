import React, { useState } from 'react';
import {
  BuildOutlined, CarryOutOutlined,
  DollarOutlined,
  FileTextOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import { Layout, Menu } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import HomeLayout from '../../Layout.jsx';

const { Sider, Content } = Layout;

export function ProjectDetails() {
  const navigate = useNavigate();
  const [navMenuItems] = useState([{
    key: 'taxonomy',
    label: 'Taxonomy',
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
        key: 'payments',
        label: 'Payments',
        icon: <DollarOutlined />,
      },
      {
        key: 'assets',
        label: 'Assets',
        icon: <CarryOutOutlined />,
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
            style={{ height: '92vh' }}
            defaultSelectedKeys={['payments']}
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
