import React, { useEffect, useState } from 'react';
import { Menu, Icon, Popover, Button } from 'antd';
import { SVG } from '../factorsComponents';
import { useLocation, NavLink } from 'react-router-dom';
import styles from './index.module.scss';
import { fetchSmartEvents } from 'Reducers/events';
import { connect } from 'react-redux';
import { fetchProjectAgents, fetchAgentInfo } from 'Reducers/agentActions';
import { fetchProjects } from 'Reducers/global';
import { getActiveDomain } from '@sentry/hub';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { APP_LAYOUT_ROUTES } from 'Routes/constants';

const { SubMenu } = Menu;

const whiteListedAccounts = [
  'baliga@factors.ai',
  'solutions@factors.ai',
  'sonali@factors.ai',
  'praveenr@factors.ai',
  'janani@factors.ai',
  'vikas@factors.ai'
];

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
  Insights: 'Build Insights',
  Sharing: 'Sharing',
  visitorIdentification: 'Visitor Identification'
};

const MapNametToLocation = {
  dashboardFilled: '/',
  analysis: '/analyse',
  profile: '/profile',
  explain: '/explain',
  attribution: '/attribution',
  configure: '/configure',
  settings: '/settings',
  setup_assist: '/welcome',
  PathAnalysis: '/path-analysis'
};

function SiderMenu({
  collapsed,
  setCollapsed,
  handleClick,
  activeProject,
  activeAgent,
  fetchSmartEvents,
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjects,
  currentProjectSettings
}) {
  const location = useLocation();
  const [openKeys, setOpenKeys] = useState([]);
  const [ShowPopOverSettings, setShowPopOverSettings] = useState(false);
  const [ShowPopOverConfigure, setShowPopOverConfigure] = useState(false);
  const [ShowPopOverProfiles, setShowPopOverProfiles] = useState(false);

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
        'Alerts'
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
      const items = [
        'general',
        'User',
        'Attribution',
        'SDK',
        'Integration',
        'Insights',
        'Sharing'
      ];
      return (
        <div className={styles.popover_content}>
          {items.map((item) => {
            if (
              item == 'Insights' &&
              !whiteListedAccounts.includes(activeAgent)
            ) {
              return null;
            }
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
    } else if (title === 'profile') {
      const items = ['People', 'Accounts', 'visitorIdentification'];

      return (
        <div className={styles.popover_content}>
          {items.map((item) => {
            return (
              <NavLink
                activeStyle={{ color: '#1890ff' }}
                exact
                to={
                  item !== 'visitorIdentification'
                    ? `/profiles/${item.toLowerCase()}`
                    : APP_LAYOUT_ROUTES.VisitorIdentificationReport.path
                }
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

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
      await fetchProjects();
      await fetchProjectAgents(activeProject?.id);
    };
    getData();
    if (location.pathname === '/configure/events') {
      fetchSmartEvents(activeProject?.id);
    }
  }, [location.pathname, activeProject]);

  const onClickAction = (key) => {
    if (key.key === 'collapse') {
      setCollapsed(!collapsed);
    } else handleClick(key);
  };

  const setIcon = (name, size = 28) => {
    let color = '#8692A3';
    if (location.pathname === MapNametToLocation[name]) {
      color = 'purple';
    }
    if (
      name == 'profile' ||
      name == 'configure' ||
      name == 'settings' ||
      name == 'PathAnalysis'
    ) {
      if (location.pathname.includes(MapNametToLocation[name])) {
        color = 'purple';
      }
    }
    if (name === 'AttributionV1') {
      if (Object.values(ATTRIBUTION_ROUTES).includes(location.pathname))
        color = 'purple';
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
      className={styles.menu}
    >
      {/* <div style={{height:}}></div> */}
      <Menu.Item
        className={styles.menuitems}
        key='/'
        icon={setIcon('dashboardFilled')}
      >
        <b>Dashboard</b>
      </Menu.Item>
      <Menu.Item
        className={styles.menuitems}
        key='/analyse'
        icon={setIcon('analysis')}
      >
        <b>Analyse</b>
      </Menu.Item>
      {collapsed ? (
        <Popover
          overlayClassName={styles.popover}
          title={false}
          visible={ShowPopOverProfiles}
          content={renderSubmenu('profile')}
          placement={'rightTop'}
          onVisibleChange={(visible) => {
            setShowPopOverProfiles(visible);
          }}
          trigger='hover'
        >
          <Menu.Item
            className={styles.menuitems}
            key='sub3'
            icon={setIcon('profile')}
          ></Menu.Item>
        </Popover>
      ) : (
        <SubMenu key='sub1' icon={setIcon('profile')} title={<b>Profiles</b>}>
          <Menu.Item className={styles.menuitems_sub} key={`/profiles/people`}>
            {MenuItems.People}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={`/profiles/accounts`}
          >
            {MenuItems.Accounts}
          </Menu.Item>
          <Menu.Item
            className={styles.menuitems_sub}
            key={APP_LAYOUT_ROUTES.VisitorIdentificationReport.path}
          >
            {MenuItems.visitorIdentification}
          </Menu.Item>
        </SubMenu>
      )}

      <Menu.Item
        className={styles.menuitems}
        key='/explain'
        icon={setIcon('explain')}
      >
        <b>Explain</b>
      </Menu.Item>

      {currentProjectSettings?.is_path_analysis_enabled && (
        <>
          <Menu.Item
            className={styles.menuitems}
            key='/path-analysis'
            icon={setIcon('PathAnalysis')}
          >
            <b>Path Analysis</b>
          </Menu.Item>
        </>
      )}
      {whiteListedAccounts.includes(activeAgent) && (
        <>
          <Menu.Item
            className={styles.menuitems}
            key='/attribution'
            icon={setIcon('AttributionV1')}
          >
            <b>Attribution</b>
          </Menu.Item>
        </>
      )}

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
        <SubMenu key='sub3' icon={setIcon('settings')} title={<b>Settings</b>}>
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
          <Menu.Item className={styles.menuitems_sub} key={`/settings/sharing`}>
            {MenuItems.Sharing}
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
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  activeAgent: state.agent?.agent_details?.email,
  currentProjectSettings: state.global.currentProjectSettings
});
export default connect(mapStateToProps, {
  fetchSmartEvents,
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjects
})(SiderMenu);
