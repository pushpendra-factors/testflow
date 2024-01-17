import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { matchPath } from 'react-router-dom';
import { getASCIISum } from 'Utils/global';
import { PathUrls } from '../../routes/pathUrls';
import { segmentColorCodes } from './appSidebar.constants';

export const checkMatchPath = (urlToMatch, internalRouteToMatch) =>
  matchPath(urlToMatch, {
    path: internalRouteToMatch,
    exact: true,
    strict: false
  });

export const isSettingsUrl = (pathname) =>
  pathname === PathUrls.SettingsGeneral ||
  pathname === PathUrls.SettingsIntegration ||
  pathname === PathUrls.SettingsSdk ||
  pathname === PathUrls.SettingsSharing ||
  pathname === PathUrls.SettingsUser ||
  pathname === PathUrls.SettingsPricing;

export const isConfigurationUrl = (pathname) =>
  pathname === PathUrls.ConfigureAlerts ||
  pathname === PathUrls.ConfigureContentGroups ||
  pathname === PathUrls.ConfigureCustomKpi ||
  pathname === PathUrls.ConfigureDataPoints ||
  pathname === PathUrls.ConfigureEvents ||
  pathname === PathUrls.ConfigureProperties ||
  pathname === PathUrls.ConfigureTouchPoints ||
  pathname === PathUrls.ConfigureEngagements ||
  pathname === PathUrls.ConfigureAttribution ||
  pathname === PathUrls.ConfigurePlans;

export const isAccountsUrl = (pathname) =>
  pathname === PathUrls.ProfileAccounts ||
  pathname === PathUrls.ProfilePeople ||
  pathname === PathUrls.VisitorIdentificationReport;

export const isReportsUrl = (pathname) =>
  pathname === PathUrls.Dashboard ||
  pathname === PathUrls.Analyse2 ||
  pathname === PathUrls.PreBuildDashboard;

export const isJourneyUrl = (pathname) =>
  pathname === PathUrls.Explain || pathname === PathUrls.PathAnalysis;

export const isAttributionsUrl = (pathname) =>
  pathname === ATTRIBUTION_ROUTES.base ||
  pathname === ATTRIBUTION_ROUTES.report ||
  pathname === ATTRIBUTION_ROUTES.reports;

export const getSegmentColorCode = (str) => {
  const asciiSum = getASCIISum(str);
  const index = asciiSum % segmentColorCodes.length;
  return segmentColorCodes[index];
};
