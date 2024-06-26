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
  setDrawerVisibleAction,
  setNewSegmentModeAction,
  toggleAccountsTab
} from 'Reducers/accountProfilesView/actions';
import { selectTimelinePayload } from 'Reducers/userProfilesView/selectors';
import { setTimelinePayloadAction } from 'Reducers/userProfilesView/actions';
import {
  isProfileAccountsUrl,
  isProfilePeopleUrl,
  isProfilesUrl,
  isSettingsUrl
} from './appSidebar.helpers';
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
  const profileActiveSegment = timelinePayload?.segment;
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);
  const activeSegment = activeAccountPayload?.segment;

  const isAllAccountsSelected = Boolean(activeSegment?.id) === false;

  const { newSegmentMode: profileNewSegmentMode } = useSelector(
    (state) => state.userProfilesView
  );

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
    dispatch(setNewSegmentModeAction(false));
    dispatch(toggleAccountsTab('accounts'));
    history.replace(PathUrls.ProfileAccounts);
  };

  const selectAllAccounts = () => {
    if (isAllAccountsSelected === false || newSegmentMode === true) {
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
            'flex flex-col gap-y-4',
            styles['sidebar-content-container'],
            {
              'pt-6': !isProfilesUrl(pathname)
            }
          )}
        >
          <div
            className={cx('flex justify-between items-center', {
              'py-2 h-12 border-b border-neutral-grey-4':
                isProfilesUrl(pathname)
            })}
          >
            <ControlledComponent controller={!isProfilesUrl(pathname)}>
              <div
                className={cx('flex gap-x-2 items-center px-3', {
                  'pl-6': sidebarTitleConfig.title === 'Dashboards',
                  'pl-4': isSettingsUrl(pathname)
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
            <ControlledComponent controller={isProfilesUrl(pathname)}>
              <ControlledComponent controller={isProfileAccountsUrl(pathname)}>
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
                    extraClass='h-8'
                  />
                </div>
              </ControlledComponent>
              <ControlledComponent controller={isProfilePeopleUrl(pathname)}>
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
        <div
          className={cx('flex justify-end', {
            'mt-2': isProfilesUrl(pathname),
            'mt-5': !isProfilesUrl(pathname)
          })}
        >
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
