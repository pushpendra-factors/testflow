import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { PathUrls } from '../../routes/pathUrls';
import { matchPath } from 'react-router-dom';

export const checkMatchPath = (urlToMatch, internalRouteToMatch) => {
  return matchPath(urlToMatch, {path: internalRouteToMatch, exact: true, strict: false});
}

export const isSettingsUrl = (pathname) => {
  return (
    pathname === PathUrls.SettingsGeneral ||
    pathname === PathUrls.SettingsIntegration ||
    pathname === PathUrls.SettingsSdk ||
    pathname === PathUrls.SettingsSharing ||
    pathname === PathUrls.SettingsUser ||
    pathname === PathUrls.SettingsPricing
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
    pathname === PathUrls.ConfigureTouchPoints ||
    pathname === PathUrls.ConfigureEngagements ||
    pathname === PathUrls.ConfigureAttribution ||
    pathname === PathUrls.ConfigurePlans
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
  return pathname === PathUrls.Dashboard || pathname === PathUrls.Analyse2 || pathname === PathUrls.PreBuildDashboard;
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
