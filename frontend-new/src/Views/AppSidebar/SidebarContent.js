import React from 'react';
import { useLocation } from 'react-router-dom';
import { PathUrls } from '../../routes/pathUrls';
import DashboardSidebar from './DashboardSidebar';
import AccountsSidebar from './AccountsSidebar';
import ProfilesSidebar from './ProfilesSidebar';
import SettingsSidebar from './SettingsSidebar';
import AlertsSidebar from './AlertsSidebar';
import { checkMatchPath, isSettingsUrl } from './appSidebar.helpers';

function SidebarContent() {
  const location = useLocation();
  const { pathname } = location;

  if (
    checkMatchPath(pathname, PathUrls.Dashboard) ||
    checkMatchPath(pathname, PathUrls.DashboardURL) ||
    checkMatchPath(pathname, PathUrls.PreBuildDashboard)
  ) {
    return <DashboardSidebar />;
  }
  if (
    checkMatchPath(pathname, PathUrls.ProfileAccounts) ||
    checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL)
  ) {
    return <AccountsSidebar />;
  }
  if (pathname === PathUrls.ProfilePeople) {
    return <ProfilesSidebar />;
  }
  if (isSettingsUrl(pathname)) {
    return <SettingsSidebar />;
  }
  if (
    checkMatchPath(pathname, PathUrls.Alerts) ||
    checkMatchPath(pathname, PathUrls.Workflows)
  ) {
    return <AlertsSidebar />;
  }
  return null;
}

export default SidebarContent;
