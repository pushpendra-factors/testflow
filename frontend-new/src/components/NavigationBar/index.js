import React from 'react';
import { Layout, Button } from 'antd';
import { LeftOutlined, RightOutlined } from '@ant-design/icons';
import { useHistory } from 'react-router-dom';
import styles from './index.module.scss';
import SiderMenu from './Menu';
import { SVG } from '../factorsComponents';

function NavigationBar(props) {
  const { Sider } = Layout;
  const history = useHistory();

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
            <Button
              type='text'
              className={`${styles.button} w-full`}
              icon={<SVG name='arrow_right'/>}
            ></Button>
          ) : (
            <Button
              type='text'
              className={`${styles.button} w-full`}
              icon={<SVG name='arrow_left'/>}
            >
              Collapse
            </Button>
          )
        }
      >
        <div>
          <SiderMenu
            collapsed={props.collapse}
            setCollapsed={props.setCollapse}
            handleClick={handleClick}
          />
        </div>
      </Sider>
    </div>
  );
}
export default NavigationBar;
