import React from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Layout } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { selectSidebarCollapsedState } from 'Reducers/global/selectors';
import { toggleSidebarCollapsedStateAction } from 'Reducers/global/actions';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { PathUrls } from 'Routes/pathUrls';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setDrawerVisibleAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import { selectTimelinePayload } from 'Reducers/userProfilesView/selectors';
import { setTimelinePayloadAction } from 'Reducers/userProfilesView/actions';
import { INITIAL_ACCOUNT_PAYLOAD } from 'Reducers/accountProfilesView';
import { checkMatchPath } from './appSidebar.helpers';
import SidebarMenuItem from './SidebarMenuItem';
import styles from './index.module.scss';
import SidebarContent from './SidebarContent';
import useSidebarTitleConfig from './hooks/useSidebarTitleConfig';

function AppSidebar() {
  const { Sider } = Layout;
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const { pathname } = location;
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));

  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);
  const activeSegment = activeAccountPayload?.segment;

  const isAllAccountsSelected = Boolean(activeSegment) === false;

  const {
    newSegmentMode: profileNewSegmentMode,
    activeSegment: profileActiveSegment
  } = useSelector((state) => state.userProfilesView);

  const isAllUsersSelected =
    timelinePayload.source === 'All' &&
    Boolean(profileActiveSegment?.name) === false;

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

  const changeAccountPayload = () => {
    dispatch(setDrawerVisibleAction(false));
    dispatch(setAccountPayloadAction(INITIAL_ACCOUNT_PAYLOAD));
    dispatch(setNewSegmentModeAction(false));
    history.replace(PathUrls.ProfileAccounts);
  };

  const selectAllAccounts = () => {
    if (isAllAccountsSelected === false) {
      changeAccountPayload();
    }
  };

  const selectAllUsers = () => {
    if (isAllUsersSelected === false) {
      dispatch(
        setTimelinePayloadAction({
          source: 'All',
          segment: {}
        })
      );
    }
  };

  return (
    <Sider
      className={cx(styles['app-sidebar'], 'fixed h-full', {
        [styles.collapsed]: isSidebarCollapsed
      })}
      onClick={isSidebarCollapsed ? handleExpand : null}
    >
      <ControlledComponent controller={isSidebarCollapsed === false}>
        <div
          className={cx(
            'flex flex-col gap-y-4 pt-6',
            styles['sidebar-content-container']
          )}
        >
          <div
            className={cx('flex justify-between items-center', {
              'pb-5 border-b border-gray-300':
                checkMatchPath(pathname, PathUrls.ProfileAccounts) ||
                checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL) ||
                checkMatchPath(pathname, PathUrls.ProfilePeople)
            })}
          >
            <ControlledComponent
              controller={
                !checkMatchPath(pathname, PathUrls.ProfileAccounts) &&
                !checkMatchPath(
                  pathname,
                  PathUrls.ProfileAccountsSegmentsURL
                ) &&
                !checkMatchPath(pathname, PathUrls.ProfilePeople)
              }
            >
              <div
                className={cx('flex gap-x-2 items-center px-3', {
                  'pl-6': sidebarTitleConfig.title === 'Dashboards'
                })}
              >
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
              controller={
                checkMatchPath(pathname, PathUrls.ProfileAccounts) ||
                checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL) ||
                checkMatchPath(pathname, PathUrls.ProfilePeople)
              }
            >
              <ControlledComponent
                controller={
                  checkMatchPath(pathname, PathUrls.ProfileAccounts) ||
                  checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL)
                }
              >
                <div className='w-11/12 pl-4'>
                  <SidebarMenuItem
                    isActive={
                      isAllAccountsSelected === true && newSegmentMode === false
                    }
                    text='All Accounts'
                    onClick={selectAllAccounts}
                    icon='regularBuilding'
                    iconColor='#F5222D'
                    iconSize={20}
                  />
                </div>
              </ControlledComponent>
              <ControlledComponent
                controller={checkMatchPath(pathname, PathUrls.ProfilePeople)}
              >
                <div className='w-11/12 pl-4'>
                  <SidebarMenuItem
                    isActive={
                      isAllUsersSelected === true &&
                      profileNewSegmentMode === false
                    }
                    text='All People'
                    onClick={selectAllUsers}
                    icon='userGroup'
                    iconColor='#FA541C'
                    iconSize={20}
                  />
                </div>
              </ControlledComponent>
            </ControlledComponent>
            <div
              role='button'
              tabIndex='0'
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
            tabIndex='-2'
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
}

export default AppSidebar;
