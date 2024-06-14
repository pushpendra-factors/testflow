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
  pathname === PathUrls.SettingsSharing ||
  pathname === PathUrls.SettingsMembers ||
  pathname === PathUrls.SettingsPricing ||
  pathname === PathUrls.SettingsPersonalProjects ||
  pathname === PathUrls.SettingsPersonalUser ||
  pathname === PathUrls.SettingsTouchpointDefinition ||
  pathname === PathUrls.SettingsCustomDefinition ||
  pathname === PathUrls.ConfigurePlans ||
  pathname === PathUrls.SettingsAttribution ||
  pathname === PathUrls.SettingsAccountScoring ||
  checkMatchPath(pathname, PathUrls.SettingsIntegrationURLID);

export const isAccountsUrl = (pathname) =>
  pathname === PathUrls.ProfileAccounts ||
  pathname === PathUrls.ProfilePeople ||
  pathname === PathUrls.VisitorIdentificationReport;

export const isReportsUrl = (pathname) =>
  pathname === PathUrls.Dashboard ||
  pathname === PathUrls.Analyse2 ||
  pathname === PathUrls.PreBuildDashboard;

export const isReportsMainUrl = (pathname) =>
  pathname === PathUrls.Explain ||
  pathname === PathUrls.PathAnalysis ||
  isReportsUrl(pathname) ||
  isAttributionsUrl(pathname);

export const isAlertsUrl = (pathname) =>
  pathname === PathUrls.Alerts || pathname === PathUrls.Workflows;

export const isCampaignsUrl = (pathname) =>
  checkMatchPath(pathname, PathUrls.FreqCap) ||
  checkMatchPath(pathname, PathUrls.FreqCapView);

export const isAttributionsUrl = (pathname) =>
  pathname === ATTRIBUTION_ROUTES.base ||
  pathname === ATTRIBUTION_ROUTES.report ||
  pathname === ATTRIBUTION_ROUTES.reports;

export const getSegmentColorCode = (str) => {
  const asciiSum = getASCIISum(str);
  const index = asciiSum % segmentColorCodes.length;
  return segmentColorCodes[index];
};
