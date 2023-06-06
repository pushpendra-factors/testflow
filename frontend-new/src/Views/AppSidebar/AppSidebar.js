import React from 'react';
import cx from 'classnames';
import { Layout } from 'antd';
import styles from './index.module.scss';
import useSidebarTitleConfig from './hooks/useSidebarTitleConfig';
import { SVG, Text } from 'Components/factorsComponents';
import SidebarContent from './SidebarContent';
import { useDispatch, useSelector } from 'react-redux';
import { selectSidebarCollapsedState } from 'Reducers/global/selectors';
import { toggleSidebarCollapsedStateAction } from 'Reducers/global/actions';

const AppSidebar = () => {
  const { Sider } = Layout;
  const dispatch = useDispatch();
  const isSidebarCollapsed = useSelector((state) =>
    selectSidebarCollapsedState(state)
  );
  const sidebarTitleConfig = useSidebarTitleConfig();

  const handleCollapse = () => {
    dispatch(toggleSidebarCollapsedStateAction(true));
  };

  const handleExpand = () => {
    dispatch(toggleSidebarCollapsedStateAction(false));
  };

  return (
    <Sider
      className={cx(styles['app-sidebar'], 'fixed h-full', {
        [styles['collapsed']]: isSidebarCollapsed
      })}
      onClick={isSidebarCollapsed ? handleExpand : null}
    >
      {isSidebarCollapsed === false && (
        <div
          className={cx(
            'flex flex-col row-gap-4 pt-6',
            styles['sidebar-content-container']
          )}
        >
          <div className='flex justify-between items-center'>
            <div className='flex col-gap-2 items-center px-3'>
              <SVG
                color={sidebarTitleConfig.iconColor}
                name={sidebarTitleConfig.icon}
              />
              <Text type='title' extraClass='mb-0' color='character-secondary'>
                {sidebarTitleConfig.title}
              </Text>
            </div>
            <div
              role='button'
              onClick={handleCollapse}
              className={cx(
                'flex justify-center items-center w-8 h-8 rounded-full',
                styles['collapsible-icon-wrapper']
              )}
            >
              <SVG name='arrow_left' color='#8C8C8C' size={20} />
            </div>
          </div>
          <SidebarContent />
        </div>
      )}
      {isSidebarCollapsed === true && (
        <div className='flex mt-5 justify-end'>
          <div
            role='button'
            onClick={handleExpand}
            className={cx(
              'flex justify-center items-center w-8 h-8 rounded-full',
              styles['collapsible-icon-wrapper']
            )}
          >
            <SVG name='arrow_right' color='#8C8C8C' size={20} />
          </div>
        </div>
      )}
    </Sider>
  );
};

export default AppSidebar;
