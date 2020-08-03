import React from 'react';
import Sidebar from './components/Sidebar';

import { Layout } from 'antd';

function App() {
  const { Content } = Layout;

  return (
    <div className="App">
      <Layout>
        <Sidebar collapsed={true} />
        <Layout className="site-layout">
          <Content>
            <div className="p-8 b-white">
              Content
            </div>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}

export default App;
