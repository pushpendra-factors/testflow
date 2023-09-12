import React from 'react';
import { useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Layout } from 'antd';
import useSidebarTitleConfig from './hooks/useSidebarTitleConfig';
import { SVG, Text } from 'Components/factorsComponents';
import { selectSidebarCollapsedState } from 'Reducers/global/selectors';
import { toggleSidebarCollapsedStateAction } from 'Reducers/global/actions';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { PathUrls } from 'Routes/pathUrls';
import SidebarContent from './SidebarContent';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setActiveSegmentAction
} from 'Reducers/accountProfilesView/actions';

const AppSidebar = () => {
  const { Sider } = Layout;
  const dispatch = useDispatch();
  const location = useLocation();
  const { pathname } = location;
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);
  const isAllAccountsSelected =
    activeAccountPayload.source === 'All' &&
    Boolean(activeAccountPayload.segment_id) === false;

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

  const selectAllAccounts = () => {
    dispatch(
      setAccountPayloadAction({
        source: 'All',
        filters: [],
        segment_id: ''
      })
    );
    dispatch(setActiveSegmentAction({}));
  };

  return (
    <Sider
      className={cx(styles['app-sidebar'], 'fixed h-full', {
        [styles['collapsed']]: isSidebarCollapsed
      })}
      onClick={isSidebarCollapsed ? handleExpand : null}
    >
      <ControlledComponent controller={isSidebarCollapsed === false}>
        <div
          className={cx(
            'flex flex-col row-gap-4 pt-6',
            styles['sidebar-content-container']
          )}
        >
          <div
            className={cx('flex justify-between items-center', {
              'pb-5 border-b border-gray-300':
                pathname === PathUrls.ProfileAccounts
            })}
          >
            <ControlledComponent
              controller={pathname !== PathUrls.ProfileAccounts}
            >
              <div className='flex col-gap-2 items-center px-3'>
                <SVG
                  color={sidebarTitleConfig.iconColor}
                  name={sidebarTitleConfig.icon}
                />
                <Text
                  type='title'
                  extraClass='mb-0'
                  color='character-secondary'
                >
                  {sidebarTitleConfig.title}
                </Text>
              </div>
            </ControlledComponent>
            <ControlledComponent
              controller={pathname === PathUrls.ProfileAccounts}
            >
              <div className='w-11/12 pl-4'>
                <SidebarMenuItem
                  isActive={
                    isAllAccountsSelected === true && newSegmentMode === false
                  }
                  text='All Accounts'
                  onClick={selectAllAccounts}
                  icon='regularBuilding'
                  iconColor={'#F5222D'}
                  iconSize={20}
                />
              </div>
            </ControlledComponent>
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
      </ControlledComponent>
      <ControlledComponent controller={isSidebarCollapsed === true}>
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
      </ControlledComponent>
    </Sider>
  );
};

export default AppSidebar;
