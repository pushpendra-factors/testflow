import { useLocation } from 'react-router-dom';
import { PathUrls } from '../../../routes/pathUrls';
import { isConfigurationUrl, isSettingsUrl } from '../appSidebar.helpers';

const useSidebarTitleConfig = () => {
  const location = useLocation();
  const { pathname } = location;

  if (pathname === PathUrls.Dashboard) {
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
  if (pathname === PathUrls.ProfileAccounts) {
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
      icon: 'settings_Filled',
      iconColor: '#8C8C8C'
    };
  }

  if (isConfigurationUrl(pathname)) {
    return {
      title: 'Configure',
      icon: 'configure_Filled',
      iconColor: '#8C8C8C'
    };
  }
  return '';
};

export default useSidebarTitleConfig;
