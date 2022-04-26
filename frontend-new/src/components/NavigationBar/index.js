import React, { useState, useEffect } from 'react';
import { Button, Layout, Menu, Popover } from 'antd';
import { LeftOutlined, RightOutlined } from '@ant-design/icons';
import { SVG } from '../factorsComponents';
import { useHistory, useLocation } from 'react-router-dom';
import styles from './index.module.scss';
import SiderMenu from './Menu';

function NavigationBar(props) {
  const { Sider } = Layout;
  const history = useHistory();
  // const location = useLocation();

  const onCollapse = () => {
    props.setCollapse(!props.collapse);
  };

  const handleClick = (e) => {
    history.push(e.key.toLowerCase());
  };

  return (
    <div>
      <Sider
        collapsedWidth={64}
        width={264}
        className={styles.sider}
        collapsible
        collapsed={props.collapse}
        onCollapse={onCollapse}
        trigger={
          props.collapse ? (
            <RightOutlined />
          ) : (
            <div className='flex items-center justify-center'>
              <LeftOutlined /> Collapse
            </div>
          )
        }
      >
        <div
        // onMouseEnter={() => setCollapsed(false)} onMouseLeave={()=>setCollapsed(true)}
        >
          <SiderMenu collapsed={props.collapse} setCollapsed={props.setCollapse} handleClick={handleClick} />
        </div>
      </Sider>
    </div>
  );
}
export default NavigationBar;
