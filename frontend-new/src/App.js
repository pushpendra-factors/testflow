import React from 'react';
import Sidebar from './components/Sidebar';

import { Layout } from 'antd';
import CoreQuery from './Views/CoreQuery';

function App() {
  const { Content } = Layout;

  return (
    <div className="App">
      <Layout>
        <Sidebar />
        <Layout className="site-layout">
          <Content>
            <div className="p-4 bg-white min-h-screen">
              <CoreQuery />
            </div>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}

export default App;
