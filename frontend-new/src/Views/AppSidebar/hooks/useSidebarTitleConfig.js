import { useLocation } from 'react-router-dom';
import { PathUrls } from '../../../routes/pathUrls';
import {
  checkMatchPath,
  isAlertsUrl,
  isCampaignsUrl,
  isSettingsUrl
} from '../appSidebar.helpers';

const useSidebarTitleConfig = () => {
  const location = useLocation();
  const { pathname } = location;

  if (
    checkMatchPath(pathname, PathUrls.Dashboard) ||
    checkMatchPath(pathname, PathUrls.DashboardURL) ||
    checkMatchPath(pathname, PathUrls.PreBuildDashboard)
  ) {
    return {
      title: 'Dashboards',
      icon: 'dashboard_Filled',
      iconColor: '#40A9FF'
    };
  }

  if (pathname === PathUrls.Analyse2) {
    return {
      title: 'Analyse',
      icon: 'analysis_Filled',
      iconColor: '#9254DE'
    };
  }
  if (
    checkMatchPath(pathname, PathUrls.ProfileAccounts) ||
    checkMatchPath(pathname, PathUrls.ProfileAccountsSegmentsURL)
  ) {
    return {
      title: 'Accounts',
      icon: 'accounts',
      iconColor: '#597EF7'
    };
  }
  if (pathname === PathUrls.ProfilePeople) {
    return {
      title: 'Profiles',
      icon: 'coloredProfile'
    };
  }
  if (isSettingsUrl(pathname)) {
    return {
      title: 'Settings',
      icon: '',
      iconColor: '#8C8C8C'
    };
  }
  if (isAlertsUrl(pathname)) {
    return {
      title: 'Automations',
      icon: 'radar',
      iconColor: '#8C8C8C'
    };
  }

  if (isCampaignsUrl(pathname)) {
    return {
      title: 'Campaign Optimizer',
      icon: '',
      iconColor: ''
    };
  }

  return '';
};

export default useSidebarTitleConfig;
