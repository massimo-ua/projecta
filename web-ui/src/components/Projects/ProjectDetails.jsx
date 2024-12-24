import React, { useState } from 'react';
import {
  BuildOutlined,
  CarryOutOutlined,
  DollarOutlined,
  FileTextOutlined,
  MenuOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import { Button, Dropdown, Grid, Layout, Menu, Space } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import HomeLayout from '../../Layout';

const { useBreakpoint } = Grid;

const { Sider, Content } = Layout;

export function ProjectDetails() {
  const navigate = useNavigate();
  const screens = useBreakpoint();
  const [ navItems ] = useState([ {
    key: 'taxonomy',
    label: 'Taxonomy',
    type: 'group',
    children: [
      {
        key: 'categories',
        label: 'Categories',
        icon: <PieChartOutlined/>,
      },
      {
        key: 'types',
        label: 'Types',
        icon: <BuildOutlined/>,
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
        icon: <FileTextOutlined/>,
      },
      {
        key: 'payments',
        label: 'Payments',
        icon: <DollarOutlined/>,
      },
      {
        key: 'assets',
        label: 'Assets',
        icon: <CarryOutOutlined/>,
      },
    ],
  } ]);

  const onClick = (e) => {
    const { key } = e;
    navigate(key);
  };

  const toDropDownItems = () => navItems.map((i) => i.children.map((i) => ({
    ...i,
    label: <a onClick={ () => onClick({ key: i.key }) }>{ i.label }</a>
  }))).flat();

  return (
    <HomeLayout>
      <Layout>
        { screens.xs
          ? (<Dropdown menu={ {
            onClick: (e) => onClick(e),
            items: toDropDownItems(),
          } }>
            <Button type="text">
              <Space>
                Menu
                <MenuOutlined/>
              </Space>
            </Button>
          </Dropdown>)
          : (<Sider width="12vw">
            <Menu
              onClick={ onClick }
              style={ { height: '92vh' } }
              defaultSelectedKeys={ [ 'payments' ] }
              mode="inline"
              items={ navItems }
            />
          </Sider>) }
        <Content>
          <Outlet/>
        </Content>
      </Layout>
    </HomeLayout>
  );
}
