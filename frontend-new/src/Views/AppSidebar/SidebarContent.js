import React from 'react';
import { useLocation } from 'react-router-dom';
import { PathUrls } from '../../routes/pathUrls';
import DashboardSidebar from './DashboardSidebar';
import AccountsSidebar from './AccountsSidebar';
import ProfilesSidebar from './ProfilesSidebar';
import SettingsSidebar from './SettingsSidebar';
import { isConfigurationUrl, isSettingsUrl } from './appSidebar.helpers';

const SidebarContent = () => {
  const location = useLocation();
  const { pathname } = location;

  if (pathname === PathUrls.Dashboard) {
    return <DashboardSidebar />;
  }
  if (pathname === PathUrls.ProfileAccounts) {
    return <AccountsSidebar />;
  }
  if (pathname === PathUrls.ProfilePeople) {
    return <ProfilesSidebar />;
  }
  if (isSettingsUrl(pathname) || isConfigurationUrl(pathname)) {
    return <SettingsSidebar />;
  }
  return null;
};

export default SidebarContent;
