import React, { useEffect } from 'react';
import './Layout.css';
import { Col, Layout, Row } from 'antd';
import { Logo } from './components/Logo';
import Logout from './components/Logout';
import { AppFooter } from './components';
import useWebsocket from './hooks/websocket';

const { Header, Footer, Content } = Layout;

export default function HomeLayout({ children }) {
  const ws = useWebsocket();

  useEffect(() => {
    if (ws) {
      ws.onMessage((data) => {
        console.log(data);
      });
    }
  }, [ws]);

  return (
    <div className="HomeLayout_container">
      <Layout style={{ width: '100%' }}>
        <Header style={{ padding: '0 10px', height: '8vh' }}>
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
