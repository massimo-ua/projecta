import React from 'react';
import './Layout.css';
import { Col, Layout, Row } from 'antd';
import { Logo } from './components/Logo';
import Logout from './components/Logout';
import { AppFooter } from './components/index.js';

const { Header, Footer, Content } = Layout;

export default function HomeLayout({ children }) {
  return (
    <div className="HomeLayout_container">
      <Layout style={{ width: '100%' }}>
        <Header style={{ padding: '0 10px' }}>
          <Row>
            <Col span={12}><Logo /></Col>
            <Col span={12} style={{ textAlign: 'right' }}><Logout /></Col>
          </Row>

        </Header>
        <Content>{children}</Content>
        <Footer style={{ height: '10vh' }}><AppFooter /></Footer>
      </Layout>
    </div>
  );
}
