import React, { useMemo } from 'react';
import cx from 'classnames';
import { Layout, Dropdown, Menu, Button } from 'antd';
import { Link, useHistory, useLocation } from 'react-router-dom';
import SearchBar from 'Components/SearchBar';
import { SVG, Text } from 'Components/factorsComponents';
import ProjectModal from 'Components/ProjectModal';
import {
  isAccountsUrl,
  isAlertsUrl,
  // isAttributionsUrl,
  isReportsMainUrl,
  // isReportsUrl,
  isSettingsUrl
} from 'Views/AppSidebar/appSidebar.helpers';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { useSelector } from 'react-redux';
import { SolutionsAccountId } from 'Routes/constants';
import { PathUrls } from '../../routes/pathUrls';
import styles from './index.module.scss';

export const getConfigureMenuItems = (email) => {
  const configureMenuItems = [
    {
      label: 'Integrations',
      url: PathUrls.SettingsIntegration
    },
    {
      label: 'Touchpoints Definitions',
      url: PathUrls.SettingsTouchpointDefinition
    },
    {
      label: 'Custom Definitions',
      url: PathUrls.SettingsCustomDefinition
    },

    {
      label: 'Attribution',
      url: PathUrls.SettingsAttribution
    }
  ];
  if (email === SolutionsAccountId) {
    configureMenuItems.push({
      label: 'Plans',
      url: PathUrls.ConfigurePlans
    });
  }
  return configureMenuItems;
};

const ProjectSettingsSubItems = [
  {
    label: 'General',
    url: PathUrls.SettingsGeneral,
    lineBreak: false
  },
  {
    label: 'Members',
    url: PathUrls.SettingsMembers,
    lineBreak: false
  },
  {
    label: 'Plans & Billings',
    url: PathUrls.SettingsPricing,
    lineBreak: false
  },
  {
    label: 'Sharing',
    url: PathUrls.SettingsSharing,
    lineBreak: false
  }
];

const PersonalSettingsSubItems = [
  {
    label: 'User Settings',
    url: PathUrls.SettingsPersonalUser,
    lineBreak: false
  },
  {
    label: 'My Projects',
    url: PathUrls.SettingsPersonalProjects
  }
];

const DataManagementSubItems = [
  {
    label: 'Integrations',
    url: PathUrls.SettingsIntegration,
    lineBreak: false
  }
];

export const settingsCategorisedMap = (email) => [
  {
    label: 'Project Settings',
    icon: 'rocket',
    items: ProjectSettingsSubItems,
    hoverable: false
  },
  {
    label: 'Personal Settings',
    icon: 'user',
    items: PersonalSettingsSubItems,
    hoverable: false
  },
  {
    label: 'Data Management',
    icon: 'database',
    items: getConfigureMenuItems(email),
    hoverable: false
  }
];

export const settingsMenuItems = [
  {
    label: 'General Settings',
    url: PathUrls.SettingsGeneral,
    lineBreak: false
  },
  {
    label: 'Users',
    url: PathUrls.SettingsMembers,
    lineBreak: true
  },
  {
    label: 'Integrations',
    url: PathUrls.SettingsIntegration,
    lineBreak: false
  },
  {
    label: 'Sharing',
    url: PathUrls.SettingsSharing,
    lineBreak: true
  },

  {
    label: 'Pricing',
    url: PathUrls.SettingsPricing,
    lineBreak: false
  }
];

const accountsMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={PathUrls.ProfileAccounts}>
        <SVG name='accounts' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Accounts
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={PathUrls.ProfilePeople}>
        <SVG name='coloredProfile' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          People
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link
        className='items-center gap-x-2'
        to={PathUrls.VisitorIdentificationReport}
      >
        <SVG name='coloredWebsiteVisitorsIdentification' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Account identification
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const reportsMainMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={PathUrls.Dashboard}>
        <SVG name='dashboard_Filled' color='#40A9FF' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Dashboards
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={PathUrls.PathAnalysis}>
        <SVG name='pathAnalysis_Filled' color='#5CDBD3' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Path Analysis
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={PathUrls.Explain}>
        <SVG name='explain_Filled' color='#D3ADF7' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Explain
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center gap-x-2' to={ATTRIBUTION_ROUTES.base}>
        <SVG name='attribution_Filled' color='#FFADD2' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Attribution
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const SettingsMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item disabled className={styles['dropdown-menu-item']}>
      <Text color='disabled' level={7} type='title' extraClass='mb-0'>
        Settings
      </Text>
    </Menu.Item>
    {settingsMenuItems.map((item) => {
      if (item?.whitelisted) {
        return null;
      }
      return (
        <>
          <Menu.Item key={item.label} className={styles['dropdown-menu-item']}>
            <Link to={item.url}>
              <Text color='black' level={7} type='title' extraClass='mb-0'>
                {item.label}
              </Text>
            </Link>
          </Menu.Item>
          {item.lineBreak === true && <hr />}
        </>
      );
    })}
  </Menu>
);

function FaHeader() {
  const { Header } = Layout;
  const location = useLocation();
  const { pathname } = location;
  const history = useHistory();

  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;
  const activeAgentUUID = agentState?.agent_details?.uuid;

  const isChecklistEnabled = useMemo(() => {
    const agent = agentState.agents.filter(
      (data) => data.uuid === activeAgentUUID
    );
    return agent[0]?.checklist_dismissed;
  }, [agentState, agentState?.agents]);

  return (
    <Header
      className={`px-6 fixed py-3 flex items-center w-full justify-between ${styles['fa-header']}`}
    >
      <div className='flex items-center w-1/3'>
        <div className='flex items-center gap-x-6'>
          <Link to={PathUrls.ProfileAccounts} id='fa-at-link--home'>
            <SVG
              name='brand'
              background='transparent'
              showBorder={false}
              size={32}
            />
          </Link>
          <div className='flex gap-x-6'>
            <Dropdown
              overlay={accountsMenu}
              overlayClassName='fa-at-overlay--accounts'
            >
              <div
                className={cx(
                  `flex cursor-pointer items-center gap-x-1 pl-2 pr-1 py-1 ${styles['header-item']}`,
                  {
                    [styles['active-header-item']]: isAccountsUrl(pathname)
                  }
                )}
                id='fa-at-link--accounts'
              >
                <Text
                  color='white'
                  level={7}
                  type='title'
                  extraClass='mb-0'
                  weight='medium'
                >
                  Accounts
                </Text>{' '}
                <SVG color='#D9D9D9' size={16} name='chevronDown' />
              </div>
            </Dropdown>
            {/* <Link
              to={PathUrls.Dashboard}
              className={cx(
                'flex items-center pl-2 pr-1 py-1 ' + styles['header-item'],
                {
                  [styles['active-header-item']]: isReportsUrl(pathname)
                }
              )}
              id='fa-at-link--reports'
            >
              <Text
                type='title'
                color='white'
                level={7}
                extraClass='mb-0'
                weight='medium'
              >
                Reports
              </Text>
            </Link> */}
            <Dropdown overlay={reportsMainMenu}>
              <div
                className={cx(
                  `flex cursor-pointer items-center gap-x-1 pl-2 pr-1 py-1 ${styles['header-item']}`,
                  {
                    [styles['active-header-item']]: isReportsMainUrl(pathname)
                  }
                )}
                id='fa-at-link--journeys'
              >
                <Text
                  color='white'
                  level={7}
                  type='title'
                  extraClass='mb-0'
                  weight='medium'
                >
                  Reports
                </Text>{' '}
                <SVG color='#D9D9D9' size={16} name='chevronDown' />
              </div>
            </Dropdown>

            {/* <Link
              to={ATTRIBUTION_ROUTES.base}
              className={cx(
                'flex items-center pl-2 pr-1 py-1 ' + styles['header-item'],
                {
                  [styles['active-header-item']]: isAttributionsUrl(pathname)
                }
              )}
              id='fa-at-link--attribution'
            >
              <Text
                type='title'
                color='white'
                level={7}
                extraClass='mb-0'
                weight='medium'
              >
                Attribution
              </Text>
            </Link> */}

            <Link
              to={`${PathUrls.Alerts}?type=realtime`}
              className={cx(
                `flex items-center pl-2 pr-1 py-1 ${styles['header-item']}`,
                {
                  [styles['active-header-item']]: isAlertsUrl(pathname)
                }
              )}
              id='fa-at-link--attribution'
            >
              <Text
                type='title'
                color='white'
                level={7}
                extraClass='mb-0'
                weight='medium'
              >
                Automations
              </Text>
            </Link>
          </div>
        </div>
      </div>
      <div className='flex w-1/2 items-center justify-center gap-x-6 text-white'>
        {isChecklistEnabled != undefined && !isChecklistEnabled ? (
          <div className='w-1/8 flex justify-end'>
            <Button
              icon={<SVG name='Stars' size={20} extraClass='-mt-1' />}
              type='link'
              size='middle'
              onClick={() => history.push(PathUrls.Checklist)}
              className={`${styles.checklistSetup}`}
            >
              Finish setup
            </Button>
          </div>
        ) : (
          <div className='flex w-1/2 gap-x-6' />
        )}
        <div className='w-1/3 flex justify-end'>
          <SearchBar placeholder='Search ⌘+K' />
        </div>
        {/* <Dropdown
          overlay={renderConfigureMenu(activeAgent)}
          placement='bottomRight'
          overlayClassName='fa-at-overlay--config'
        >
          <div
            className={cx(
              `cursor-pointer ${styles['header-item']} ${styles['header-item-circle']}`,
              {
                [styles['active-header-item']]: isConfigurationUrl(pathname),
                [styles['active-header-item-circle']]:
                  isConfigurationUrl(pathname)
              }
            )}
            id='fa-at-dropdown--config'
          >
            <SVG color='#F0F0F0' size={16} name='config' />
          </div>
        </Dropdown> */}
        {/* <Dropdown
          placement='bottomRight'
          overlayClassName='fa-at-overlay--settings'
          overlay={SettingsMenu}
        > */}
        <div
          className={cx(
            `cursor-pointer ${styles['header-item']} ${styles['header-item-circle']}`,
            {
              [styles['active-header-item']]: isSettingsUrl(pathname),
              [styles['active-header-item-circle']]: isSettingsUrl(pathname)
            }
          )}
          onClick={() => history.push(PathUrls.SettingsGeneral)}
          id='fa-at-dropdown--settings'
        >
          <SVG color='#F0F0F0' size={20} name='settings' />
        </div>
        {/* </Dropdown> */}
        <ProjectModal />
      </div>
    </Header>
  );
}

export default FaHeader;
