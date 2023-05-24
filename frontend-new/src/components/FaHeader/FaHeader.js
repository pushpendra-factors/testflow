import React from 'react';
import cx from 'classnames';
import { Layout, Dropdown, Menu } from 'antd';
import { Link, useLocation } from 'react-router-dom';
import SearchBar from 'Components/SearchBar';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { PathUrls } from '../../routes/pathUrls';
import ProjectModal from 'Components/ProjectModal';
import {
  isAccountsUrl,
  isJourneyUrl,
  isReportsUrl
} from 'Views/AppSidebar/appSidebar.helpers';

export const configureMenuItems = [
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
    url: PathUrls.SettingsUser,
    lineBreak: true
  },
  {
    label: 'Attributions',
    url: PathUrls.SettingsAttribution,
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
    label: 'Build Insights',
    url: PathUrls.SettingsInsights,
    lineBreak: true
  },

  {
    label: 'Sharing',
    url: PathUrls.SettingsSharing,
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
          Visitor identification
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const journeyMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.PathAnalysis}>
        <SVG name='pathAnalysis' color='#73D13D' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Path Analysis
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.Explain}>
        <SVG name='explain' color='#FFC53D' />
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
      <Link className='items-center col-gap-2' to={PathUrls.Dashboard}>
        <SVG name='dashboard' color={'#40A9FF'} />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Dashboards
        </Text>
      </Link>
    </Menu.Item>
    <Menu.Item className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.Analyse2}>
        <SVG name='analysis' color='#9254DE' />
        <Text color='black' level={7} type='title' extraClass='mb-0'>
          Analyse
        </Text>
      </Link>
    </Menu.Item>
  </Menu>
);

const ConfigureMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item disabled className={styles['dropdown-menu-item']}>
      <Link to={PathUrls.Dashboard}>
        <Text color='disabled' level={7} type='title' extraClass='mb-0'>
          Configure
        </Text>
      </Link>
    </Menu.Item>
    {configureMenuItems.map((item) => {
      return (
        <Menu.Item key={item.label} className={styles['dropdown-menu-item']}>
          <Link to={item.url}>
            <Text color='black' level={7} type='title' extraClass='mb-0'>
              {item.label}
            </Text>
          </Link>
        </Menu.Item>
      );
    })}
  </Menu>
);

const settingsMenu = (
  <Menu className={styles['dropdown-menu']}>
    <Menu.Item disabled className={styles['dropdown-menu-item']}>
      <Link className='items-center col-gap-2' to={PathUrls.Dashboard}>
        <Text color='disabled' level={7} type='title' extraClass='mb-0'>
          Settings
        </Text>
      </Link>
    </Menu.Item>
    {settingsMenuItems.map((item) => {
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

  return (
    <Header
      className={`px-6 fixed py-3 flex items-center w-full justify-between ${styles['fa-header']}`}
    >
      <div className={'flex items-center w-1/3'}>
        <div className='flex items-center col-gap-6'>
          <Link to={PathUrls.Dashboard}>
            <SVG
              name={'brand'}
              background='transparent'
              showBorder={false}
              size={32}
            />
          </Link>
          <div className='flex col-gap-6'>
            <Dropdown overlay={accountsMenu}>
              <div
                className={cx('flex cursor-pointer items-center col-gap-1', {
                  [styles['active-header-item']]: isAccountsUrl(pathname)
                })}
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
            <Dropdown overlay={reportsMenu}>
              <div
                className={cx('flex cursor-pointer items-center col-gap-1', {
                  [styles['active-header-item']]: isReportsUrl(pathname)
                })}
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
            <Dropdown overlay={journeyMenu}>
              <div
                className={cx('flex cursor-pointer items-center col-gap-1', {
                  [styles['active-header-item']]: isJourneyUrl(pathname)
                })}
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
          </div>
        </div>
      </div>
      <div className='w-1/3 flex justify-center'>
        <SearchBar />
      </div>
      <div className='flex w-1/3 items-center justify-end col-gap-6 text-white'>
        <Dropdown overlay={ConfigureMenu} placement='bottomRight'>
          <div className='cursor-pointer'>
            <SVG color='#F0F0F0' size={20} name='controls' />
          </div>
        </Dropdown>
        <Dropdown placement='bottomRight' overlay={settingsMenu}>
          <div className='cursor-pointer'>
            <SVG color='#F0F0F0' size={20} name='settings' />
          </div>
        </Dropdown>

        <ProjectModal />
      </div>
    </Header>
  );
}

export default FaHeader;
