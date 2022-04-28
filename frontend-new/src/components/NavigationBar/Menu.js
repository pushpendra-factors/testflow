import React, { useState } from 'react';
import { Menu, Icon } from 'antd';
import { SVG } from '../factorsComponents';
import { useLocation, NavLink } from 'react-router-dom';
import styles from './index.module.scss';

const { SubMenu } = Menu;

const MenuItems = {
  generalSettings: 'General Settings',
  SDK: 'Javascript SDK',
  Users: 'Users',
  Integrations: 'Integrations',
  EventAlias: 'Event Alias',
  Events: 'Events',
  Properties: 'Properties',
  MarketingInteractions: 'Marketing Touchpoints',
  ContentGroups: 'Content Groups',
  Touchpoints: 'Touchpoints',
  Attributions: 'Attributions',
  CustomKPI: 'Custom KPIs',
  ExplainDP: 'Top Events and Properties',
  TargetGoals: 'Target/Goals',
  People: 'People',
  Accounts: 'Accounts',
  Campaigns: 'Campaigns',
  Pages: 'Pages',
  Alerts: 'Alerts',
};

const MapNametToLocation = {
  dashboard: '/',
  corequery: '/analyse',
  profile: '/profile',
  key: '/explain',
  attribution: '/attribution',
  configure: '/configure',
  settings: '/settings',
  Emoji: '/welcome',
};

const SiderMenu = ({ collapsed, setCollapsed, handleClick }) => {
  const location = useLocation();
  const [openKeys, setOpenKeys] = useState([]);
  const submenuKeys = ['sub1', 'sub2', 'sub3'];

  const handleOpenChange = (items) => {
    const latestOpenKey = items.find((key) => openKeys.indexOf(key) === -1);
    if (collapsed) {
      setOpenKeys([]);
    } else if (submenuKeys.indexOf(latestOpenKey) === -1) {
      setOpenKeys(items);
    } else {
      setOpenKeys(latestOpenKey ? [latestOpenKey] : []);
    }
  };

  const onClickAction = (key) => {
    if (key.key === 'collapse') {
      setCollapsed(!collapsed);
    } else handleClick(key);
  };

  const setIcon = (name, size = 24) => {
    let color;
    if (location.pathname === MapNametToLocation[name]) {
      color = 'purple';
    }
    if (name == 'profile' || name == 'configure' || name == 'settings') {
      if (location.pathname.includes(MapNametToLocation[name])) {
        color = 'purple';
      }
    }
    return (
      <span className='anticon'>
        <SVG name={name} size={size} color={color} />
      </span>
    );
  };

  return (
    <Menu
      openKeys={openKeys}
      defaultSelectedKeys={['/']}
      selectedKeys={[location.pathname]}
      mode='inline'
      onOpenChange={handleOpenChange}
      onClick={onClickAction}
      style={{ background: '#f0f2f5' }}
      className={styles.menu}
    >
      {/* <div style={{height:}}></div> */}
      <Menu.Item key='/' icon={setIcon('dashboard')}>
        <b>Dashboard</b>
      </Menu.Item>
      <Menu.Item key='/analyse' icon={setIcon('corequery')}>
        <b>Analyse</b>
      </Menu.Item>
      {/* <SubMenu
        onTitleClick={() => {
          if (collapsed) {
            setCollapsed(false);
          }
        }}
        key='sub1'
        icon={setIcon('profile')}
        title={<b>Profiles</b>}
      >
        <Menu.Item key={`/profiles/people`}>{MenuItems.People}</Menu.Item>
        <Menu.Item key={`/profiles/accounts`}>{MenuItems.Accounts}</Menu.Item>
        <Menu.Item key={`/profiles/campaigns`}>{MenuItems.Campaigns}</Menu.Item>
        <Menu.Item key={`/profiles/pages`}>{MenuItems.Pages}</Menu.Item>
      </SubMenu> */}
      <Menu.Item key='/explain' icon={setIcon('key')}>
        <b>Explain</b>
      </Menu.Item>
      {/* <Menu.Item key='/attribution' icon={setIcon('attribution')}>
        <b>Attributions</b>
      </Menu.Item> */}
      <SubMenu
        onTitleClick={() => {
          if (collapsed) {
            setCollapsed(false);
          }
        }}
        key='sub2'
        icon={setIcon('configure')}
        title={<b>Configure</b>}
      >
        <Menu.Item key={`/configure/events`}>{MenuItems.Events}</Menu.Item>
        <Menu.Item key={`/configure/properties`}>
          {MenuItems.Properties}
        </Menu.Item>
        <Menu.Item key={`/configure/contentgroups`}>
          {MenuItems.ContentGroups}
        </Menu.Item>
        <Menu.Item key={`/configure/touchpoints`}>
          {MenuItems.Touchpoints}
        </Menu.Item>
        <Menu.Item key={`/configure/customkpi`}>
          {MenuItems.CustomKPI}
        </Menu.Item>
        {/* <Menu.Item key={`/configure/goals`}>{MenuItems.TargetGoals}</Menu.Item> */}
        <Menu.Item key={`/configure/explaindp`}>
          {MenuItems.ExplainDP}
        </Menu.Item>
        <Menu.Item key={`/configure/alerts`}>{MenuItems.Alerts}</Menu.Item>
      </SubMenu>
      <SubMenu
        onTitleClick={() => {
          if (collapsed) {
            setCollapsed(false);
          }
        }}
        key='sub3'
        icon={setIcon('settings')}
        title={<b>Settings</b>}
      >
        <Menu.Item key={`/settings/general`}>
          {MenuItems.generalSettings}
        </Menu.Item>
        <Menu.Item key={`/settings/user`}>{MenuItems.Users}</Menu.Item>
        <Menu.Item key={`/settings/attribution`}>
          {MenuItems.Attributions}
        </Menu.Item>
        <Menu.Item key={`/settings/sdk`}>{MenuItems.SDK}</Menu.Item>
        <Menu.Item key={`/settings/integration`}>
          {MenuItems.Integrations}
        </Menu.Item>
      </SubMenu>
      <Menu.Item
        className={styles.menu_assist}
        key='/welcome'
        icon={setIcon('Emoji')}
      >
        <b>Setup Assist</b>
      </Menu.Item>
      <Menu.Item
        className={styles.menu_collapse}
        key='collapse'
        icon={setIcon(collapsed ? 'arrow_right' : 'arrow_left')}
      >
        <b>{collapsed ? 'Expand' : 'Collapse'}</b>
      </Menu.Item>
      <div style={{ height: '120px' }}></div>
    </Menu>
  );
};

export default SiderMenu;
