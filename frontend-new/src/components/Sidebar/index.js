import React from 'react';
import { Layout, Menu } from 'antd';
import styles from './index.module.scss';

function Sidebar() {
  const { Sider } = Layout;

  return (
    <Sider trigger={null} collapsible collapsed={true}>
      <div className={styles.logo} >
        <img src="./assets/icons/factors.png" alt="Factors.ai" />
      </div>
      <Menu className="menu-items" theme="dark" mode="inline" defaultSelectedKeys={['1']}>
        <Menu.Item key="1" icon={<img className="anticon" src="./assets/icons/home.svg" alt="Home" />}>
        </Menu.Item>
        <Menu.Item key="2" icon={<img className="anticon" src="./assets/icons/core-query-white.png" alt="Core Query" />}>
          nav 2
        </Menu.Item>
      </Menu>
    </Sider>
  )
}

export default Sidebar;