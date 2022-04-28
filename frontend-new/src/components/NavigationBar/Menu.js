import React, { useState } from 'react';
import { Menu, Icon, Popover, Button } from 'antd';
import { SVG } from '../factorsComponents';
import { useLocation, NavLink } from 'react-router-dom';
import styles from './index.module.scss';

const { SubMenu } = Menu;

const MenuItems = {
  general: 'General Settings',
  SDK: 'Javascript SDK',
  User: 'Users',
  Integration: 'Integrations',
  EventAlias: 'Event Alias',
  Events: 'Events',
  Properties: 'Properties',
  MarketingInteractions: 'Marketing Touchpoints',
  ContentGroups: 'Content Groups',
  Touchpoints: 'Touchpoints',
  Attribution: 'Attributions',
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
  setup_assist: '/welcome',
};

function SiderMenu({ collapsed, setCollapsed, handleClick }) {
  const location = useLocation();
  const [openKeys, setOpenKeys] = useState([]);
  const [ShowPopOverSettings, setShowPopOverSettings] = useState(false);
  const [ShowPopOverConfigure, setShowPopOverConfigure] = useState(false);

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

  const renderSubmenu = (title) => {
    if (title === 'configure') {
      const items = [
        'Events',
        'Properties',
        'ContentGroups',
        'Touchpoints',
        'CustomKPI',
        'ExplainDP',
        'Alerts',
      ];
      return (
        <div className={styles.popover_content}>
          {items.map((item) => {
            return (
              <NavLink
                activeStyle={{ color: '#1890ff' }}
                exact
                to={`/configure/${item.toLowerCase()}`}
                onClick={() => setShowPopOverConfigure(false)}
              >
                {MenuItems[item]}
              </NavLink>
            );
          })}
        </div>
      );
    } else if (title === 'settings') {
      const items = ['general', 'User', 'Attribution', 'SDK', 'Integration'];
      return (
        <div className={styles.popover_content}>
          {items.map((item) => {
            return (
              <NavLink
                activeStyle={{ color: '#1890ff' }}
                exact
                to={`/settings/${item.toLowerCase()}`}
                onClick={() => setShowPopOverSettings(false)}
              >
                {MenuItems[item]}
              </NavLink>
            );
          })}
        </div>
      );
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
      <Menu.Item
        className={styles.menuitems}
        key='/'
        icon={setIcon('dashboard')}
      >
        <b>Dashboard</b>
      </Menu.Item>
      <Menu.Item
        className={styles.menuitems}
        key='/analyse'
        icon={setIcon('corequery')}
      >
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
        <Menu.Item className={styles.menuitems} key={`/profiles/people`}>
          {MenuItems.People}
        </Menu.Item>
        <Menu.Item className={styles.menuitems} key={`/profiles/accounts`}>
          {MenuItems.Accounts}
        </Menu.Item>
        <Menu.Item className={styles.menuitems} key={`/profiles/campaigns`}>
          {MenuItems.Campaigns}
        </Menu.Item>
        <Menu.Item className={styles.menuitems} key={`/profiles/pages`}>
          {MenuItems.Pages}
        </Menu.Item>
      </SubMenu> */}
      <Menu.Item
        className={styles.menuitems}
        key='/explain'
        icon={setIcon('key')}
      >
        <b>Explain</b>
      </Menu.Item>
      {/* <Menu.Item
        className={styles.menuitems}
        key='/attribution'
        icon={setIcon('attribution')}
      >
        <b>Attributions</b>
      </Menu.Item> */}

      {collapsed ? (
        <Popover
          overlayClassName={styles.popover}
          title={false}
          visible={ShowPopOverConfigure}
          content={renderSubmenu('configure')}
          placement={'rightTop'}
          onVisibleChange={(visible) => {
            setShowPopOverConfigure(visible);
          }}
          trigger='hover'
        >
          <Menu.Item
            className={styles.menuitems}
            key='sub2'
            icon={setIcon('configure')}
          ></Menu.Item>
        </Popover>
      ) : (
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
          <Menu.Item className={styles.menuitems_sub} key={`/configure/events`}>
            {MenuItems.Events}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/configure/properties`}
          >
            {MenuItems.Properties}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/configure/contentgroups`}
          >
            {MenuItems.ContentGroups}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/configure/touchpoints`}
          >
            {MenuItems.Touchpoints}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/configure/customkpi`}
          >
            {MenuItems.CustomKPI}
          </Menu.Item>
          {/* <Menu.Item className={styles.menuitems} key={`/configure/goals`}>
            {MenuItems.TargetGoals}
          </Menu.Item> */}
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/configure/explaindp`}
          >
            {MenuItems.ExplainDP}
          </Menu.Item>
          <Menu.Item className={styles.menuitems_sub} key={`/configure/alerts`}>
            {MenuItems.Alerts}
          </Menu.Item>
        </SubMenu>
      )}
      {collapsed ? (
        <Popover
          overlayClassName={styles.popover}
          title={false}
          visible={ShowPopOverSettings}
          content={renderSubmenu('settings')}
          placement={'rightTop'}
          onVisibleChange={(visible) => {
            setShowPopOverSettings(visible);
          }}
          trigger='hover'
        >
          <Menu.Item
            className={styles.menuitems}
            key='sub3'
            icon={setIcon('settings')}
          ></Menu.Item>
        </Popover>
      ) : (
        <SubMenu
          key='sub3'
          icon={setIcon('settings')}
          title={
            <span style={{ paddingLeft: 0 }}>
              <b>Settings</b>
            </span>
          }
        >
          <Menu.Item className={styles.menuitems_sub} key={`/settings/general`}>
            {MenuItems.general}
          </Menu.Item>
          <Menu.Item className={styles.menuitems_sub} key={`/settings/user`}>
            {MenuItems.User}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/settings/attribution`}
          >
            {MenuItems.Attribution}
          </Menu.Item>
          <Menu.Item className={styles.menuitems_sub} key={`/settings/sdk`}>
            {MenuItems.SDK}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/settings/integration`}
          >
            {MenuItems.Integration}
          </Menu.Item>
        </SubMenu>
      )}
      <Menu.Item
        className={styles.menu_assist}
        key='/welcome'
        icon={setIcon('setup_assist')}
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
}

export default SiderMenu;
