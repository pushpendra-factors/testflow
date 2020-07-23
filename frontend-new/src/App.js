import React, { useState } from 'react';
import styles from './App.module.scss';

import { Layout } from 'antd';
import {
  MenuUnfoldOutlined,
  MenuFoldOutlined,
} from '@ant-design/icons';
import Sidebar from './components/Sidebar';

function App() {

  const { Header, Content } = Layout;

  const [collapsed, setCollapsed] = useState(true);

  const toggle = () => {
    setCollapsed(currState => {
      return !currState;
    })
  };

  return (
    <div className="App">
      <Layout>
        <Sidebar collapsed={collapsed} />
        <Layout className="site-layout">
          <Header className={styles.siteLayoutBackground} style={{ padding: 0 }}>
            {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
              className: styles.trigger,
              onClick: toggle,
            })}
          </Header>
          <Content
            className={styles.siteLayoutBackground}
            style={{
              margin: '24px 16px',
              padding: 24,
              minHeight: 280,
            }}
          >
            Content
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}

export default App;
