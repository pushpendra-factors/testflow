import React from 'react';
import { useLocation } from 'react-router-dom';
import { PathUrls } from '../../routes/pathUrls';
import DashboardSidebar from './DashboardSidebar';
import PreBuildDashboardSidebar from './PreBuildDashboardSidebar';
import AccountsSidebar from './AccountsSidebar';
import ProfilesSidebar from './ProfilesSidebar';
import SettingsSidebar from './SettingsSidebar';
import { checkMatchPath, isConfigurationUrl, isSettingsUrl } from './appSidebar.helpers';

const SidebarContent = () => {
  const location = useLocation();
  const { pathname } = location;

  if (checkMatchPath(pathname, PathUrls.Dashboard) || checkMatchPath(pathname, PathUrls.DashboardURL)) {
    return <DashboardSidebar />;
  }
  if (pathname === PathUrls.PreBuildDashboard) {
    return <PreBuildDashboardSidebar />;
  }
  if(checkMatchPath(pathname, PathUrls.ProfileAccounts) || checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL)) {
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
