import React from 'react';
import cx from 'classnames';
import { Layout, Dropdown, Menu, Button } from 'antd';
import { Link, useLocation } from 'react-router-dom';
import SearchBar from 'Components/SearchBar';
import { SVG, Text } from 'Components/factorsComponents';
import ProjectModal from 'Components/ProjectModal';
import {
  isAccountsUrl,
  isAttributionsUrl,
  isConfigurationUrl,
  isJourneyUrl,
  isReportsUrl,
  isSettingsUrl
} from 'Views/AppSidebar/appSidebar.helpers';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { useSelector } from 'react-redux';
import { SolutionsAccountId } from 'Routes/constants';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import { PathUrls } from '../../routes/pathUrls';
import styles from './index.module.scss';
import { featureLock } from '../../routes/feature';

export const getConfigureMenuItems = (email) => {
  const configureMenuItems = [
    {
      label: 'Events',
      url: PathUrls.ConfigureEvents
    },
    {
      label: 'Properties',
      url: PathUrls.ConfigureProperties
    },
    {
      label: 'Content Groups',
      url: PathUrls.ConfigureContentGroups
    },
    {
      label: 'Touchpoints',
      url: PathUrls.ConfigureTouchPoints
    },
    {
      label: 'Custom KPIs',
      url: PathUrls.ConfigureCustomKpi
    },
    {
      label: 'Top Events and Properties',
      url: PathUrls.ConfigureDataPoints
    },
    {
      label: 'Alerts',
      url: PathUrls.ConfigureAlerts
    },
    {
      label: 'Engagements',
      url: PathUrls.ConfigureEngagements
    },
    {
      label: 'Attribution',
      url: PathUrls.ConfigureAttribution
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

export const settingsMenuItems = [
  {
    label: 'General Settings',
    url: PathUrls.SettingsGeneral,
    lineBreak: false
  },
  {
    label: 'Users',
    url: PathUrls.SettingsUser,
    lineBreak: true
  },
  {
    label: 'Integrations',
    url: PathUrls.SettingsIntegration,
    lineBreak: false
  },
  {
    label: 'Javascript SDK',
    url: PathUrls.SettingsSdk,
    lineBreak: true
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
      <Link className='items-center col-gap-2' to={PathUrls.ProfileAccounts}>
        <SVG name='accounts' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Accounts
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.ProfilePeople}>
        <SVG name='coloredProfile' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          People
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link
        className='items-center col-gap-2'
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

const journeyMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.PathAnalysis}>
        <SVG name='pathAnalysis_Filled' color='#73D13D' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Path Analysis
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.Explain}>
        <SVG name='explain_Filled' color='#FFC53D' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Explain
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const reportsMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.PreBuildDashboard}>
        <SVG name='QuickBoard' color='#36CFC9' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Quick Boards
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.Dashboard}>
        <SVG name='dashboard_Filled' color='#40A9FF' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Dashboards
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const renderConfigureMenu = (email) => (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item disabled className={styles['dropdown-menu-item']}>
      <Text color='disabled' level={7} type='title' extraClass='mb-0'>
        Configure
      </Text>
    </Menu.Item>
    {getConfigureMenuItems(email).map((item) => (
      <Menu.Item key={item.label} className={styles['dropdown-menu-item']}>
        <Link to={item.url}>
          <Text color='black' level={7} type='title' extraClass='mb-0'>
            {item.label}
          </Text>
        </Link>
      </Menu.Item>
    ))}
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

  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;

  const { isFeatureLocked: isWebAnalyticsLocked } = useFeatureLock(
    FEATURES.FEATURE_WEB_ANALYTICS_DASHBOARD
  );

  return (
    <Header
      className={`px-6 fixed py-3 flex items-center w-full justify-between ${styles['fa-header']}`}
    >
      <div className='flex items-center w-1/3'>
        <div className='flex items-center col-gap-6'>
          <Link to={PathUrls.ProfileAccounts} id='fa-at-link--home'>
            <SVG
              name='brand'
              background='transparent'
              showBorder={false}
              size={32}
            />
          </Link>
          <div className='flex col-gap-6'>
            <Dropdown
              overlay={accountsMenu}
              overlayClassName='fa-at-overlay--accounts'
            >
              <div
                className={cx(
                  'flex cursor-pointer items-center col-gap-1 pl-2 pr-1 py-1',
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
            {isWebAnalyticsLocked ? (
              <Link
                to={PathUrls.Dashboard}
                className={cx('flex items-center pl-2 pr-1 py-1', {
                  [styles['active-header-item']]: isReportsUrl(pathname)
                })}
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
              </Link>
            ) : (
              <Dropdown overlay={reportsMenu}>
                <div
                  className={cx(
                    'flex cursor-pointer items-center col-gap-1 pl-2 pr-1 py-1',
                    {
                      [styles['active-header-item']]: isReportsUrl(pathname)
                    }
                  )}
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
            )}

            <Dropdown overlay={journeyMenu}>
              <div
                className={cx(
                  'flex cursor-pointer items-center col-gap-1 pl-2 pr-1 py-1',
                  {
                    [styles['active-header-item']]: isJourneyUrl(pathname)
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
                  Journeys
                </Text>{' '}
                <SVG color='#D9D9D9' size={16} name='chevronDown' />
              </div>
            </Dropdown>

            <Link
              to={ATTRIBUTION_ROUTES.base}
              className={cx('flex items-center pl-2 pr-1 py-1', {
                [styles['active-header-item']]: isAttributionsUrl(pathname)
              })}
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
            </Link>
          </div>
        </div>
      </div>
      <div className='ml-4 w-1/3 flex justify-center'>
        <SearchBar />
      </div>
      <div>
        <Button
          icon={<SVG name='Stars' />}
          type='link'
          href={PathUrls.Checklist}
          className={`${styles.checklistSetup}`}
        >
          Finish the setup
        </Button>
      </div>
      <div className='flex w-1/3 items-center justify-end col-gap-6 text-white'>
        <Dropdown
          overlay={renderConfigureMenu(activeAgent)}
          placement='bottomRight'
          overlayClassName='fa-at-overlay--config'
        >
          <div
            className={cx('cursor-pointer', {
              [styles['active-header-item']]: isConfigurationUrl(pathname),
              [styles['active-header-item-circle']]:
                isConfigurationUrl(pathname)
            })}
            id='fa-at-dropdown--config'
          >
            <SVG color='#F0F0F0' size={16} name='config' />
          </div>
        </Dropdown>
        <Dropdown
          placement='bottomRight'
          overlayClassName='fa-at-overlay--settings'
          overlay={SettingsMenu}
        >
          <div
            className={cx('cursor-pointer', {
              [styles['active-header-item']]: isSettingsUrl(pathname),
              [styles['active-header-item-circle']]: isSettingsUrl(pathname)
            })}
            id='fa-at-dropdown--settings'
          >
            <SVG color='#F0F0F0' size={20} name='settings' />
          </div>
        </Dropdown>

        <ProjectModal />
      </div>
    </Header>
  );
}

export default FaHeader;
