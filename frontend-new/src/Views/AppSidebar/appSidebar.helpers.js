import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { PathUrls } from '../../routes/pathUrls';

export const isSettingsUrl = (pathname) => {
  return (
    pathname === PathUrls.SettingsAttribution ||
    pathname === PathUrls.SettingsGeneral ||
    pathname === PathUrls.SettingsInsights ||
    pathname === PathUrls.SettingsIntegration ||
    pathname === PathUrls.SettingsSdk ||
    pathname === PathUrls.SettingsSharing ||
    pathname === PathUrls.SettingsUser
  );
};

export const isConfigurationUrl = (pathname) => {
  return (
    pathname === PathUrls.ConfigureAlerts ||
    pathname === PathUrls.ConfigureContentGroups ||
    pathname === PathUrls.ConfigureCustomKpi ||
    pathname === PathUrls.ConfigureDataPoints ||
    pathname === PathUrls.ConfigureEvents ||
    pathname === PathUrls.ConfigureProperties ||
    pathname === PathUrls.ConfigureTouchPoints
  );
};

export const isAccountsUrl = (pathname) => {
  return (
    pathname === PathUrls.ProfileAccounts ||
    pathname === PathUrls.ProfilePeople ||
    pathname === PathUrls.VisitorIdentificationReport
  );
};

export const isReportsUrl = (pathname) => {
  return pathname === PathUrls.Dashboard || pathname === PathUrls.Analyse2;
};

export const isJourneyUrl = (pathname) => {
  return pathname === PathUrls.Explain || pathname === PathUrls.PathAnalysis;
};

export const isAttributionsUrl = (pathname) => {
  return (
    pathname === ATTRIBUTION_ROUTES.base ||
    pathname === ATTRIBUTION_ROUTES.report ||
    pathname === ATTRIBUTION_ROUTES.reports
  );
};
