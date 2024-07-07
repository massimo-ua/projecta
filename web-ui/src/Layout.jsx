import React from 'react';
import './Layout.css';
import { Layout } from 'antd';
import { Logo } from './components/Logo';

const { Header, Footer, Content } = Layout;

export default function HomeLayout({ children }) {
  return (
    <div className="HomeLayout_container">
      <Layout style={{ width: '100%' }}>
        <Header style={{ padding: '0 10px' }}><Logo /></Header>
        <Content>{children}</Content>
        <Footer style={{ height: '10vh' }}>Footer</Footer>
      </Layout>
    </div>
  );
}
