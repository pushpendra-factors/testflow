import lazyWithRetry from 'Utils/lazyWithRetry';

import Welcome from 'Views/Settings/SetupAssist/Welcome/index';
import DashboardTemplates from 'Views/DashboardTemplates/index';
import AttributionSettings from 'Views/Settings/ProjectSettings/AttributionSettings';
import BasicSettings from 'Views/Settings/ProjectSettings/BasicSettings';
import SDKSettings from 'Views/Settings/ProjectSettings/SDKSettings';
import UserSettings from 'Views/Settings/ProjectSettings/UserSettings';
import IntegrationSettings from 'Views/Settings/ProjectSettings/IntegrationSettings';
import Sharing from 'Views/Settings/ProjectSettings/Sharing';
import Events from 'Views/Settings/ProjectConfigure/Events';
import InsightsSettings from 'Views/Settings/ProjectSettings/InsightsSettings';
import PropertySettings from 'Views/Settings/ProjectConfigure/PropertySettings';
import ContentGroups from 'Views/Settings/ProjectConfigure/ContentGroups';
import CustomKPI from 'Views/Settings/ProjectConfigure/CustomKPI';
import Alerts from 'Views/Settings/ProjectSettings/Alerts';
import ExplainDataPoints from 'Views/Settings/ProjectConfigure/ExplainDataPoints';
import UserProfiles from 'Components/Profile/UserProfiles';
import AccountProfiles from 'Components/Profile/AccountProfiles';
import Touchpoints from 'Views/Settings/ProjectConfigure/Touchpoints';
import AppLayout from 'Views/AppLayout';
import OnBoard from 'Views/Settings/SetupAssist/Welcome/OnboardFlow';
import { PathUrls } from './pathUrls';

const Login = lazyWithRetry(() => import('../Views/Pages/Login'));
const ForgotPassword = lazyWithRetry(() =>
  import('../Views/Pages/ForgotPassword')
);
const ResetPassword = lazyWithRetry(() =>
  import('../Views/Pages/ResetPassword')
);
const SignUp = lazyWithRetry(() => import('../Views/Pages/SignUp'));
const Activate = lazyWithRetry(() => import('../Views/Pages/Activate'));
const Templates = lazyWithRetry(() =>
  import('../Views/CoreQuery/Templates/ResultsPage')
);
// const AppLayout = lazyWithRetry(() => import('../Views/AppLayout'));

const FactorsInsightsNew = lazyWithRetry(() =>
  import('../Views/Factors/FactorsInsightsNew')
);
const FactorsInsightsOld = lazyWithRetry(() =>
  import('../Views/Factors/FactorsInsightsOld')
);
const CoreQuery = lazyWithRetry(() => import('../Views/CoreQuery'));
const Dashboard = lazyWithRetry(() => import('../Views/Dashboard'));
const Factors = lazyWithRetry(() => import('../Views/Factors'));
const VisitorIdentificationReportComponent = lazyWithRetry(() =>
  import('../features/6signal-report/ui')
);
const SixSignalReportRedirection = lazyWithRetry(() =>
  import('../features/6signal-report/ui/SixSignalRedirection')
);

const componentsLib = lazyWithRetry(() => import('../Views/componentsLib'));

export const APP_ROUTES = {
  Signup: {
    path: '/signup',
    Component: SignUp,
    exact: true
  },
  Activate: {
    path: '/activate',
    Component: Activate,
    exact: true
  },
  SetPassword: {
    path: '/setpassword',
    Component: ResetPassword,
    exact: true
  },
  ForgotPassword: {
    path: '/forgotpassword',
    Component: ForgotPassword,
    exact: true
  },
  Login: {
    title: 'Login',
    path: '/login',
    Component: Login,
    exact: true
  },
  Templates: {
    path: '/templates',
    Component: Templates,
    Private: true,
    exact: true
  },
  APPLayout: {
    path: '/',
    Component: AppLayout
  }
};

export const APP_LAYOUT_ROUTES = {
  Dashboard: {
    title: 'Dashboard',
    path: PathUrls.Dashboard,
    Component: Dashboard,
    exact: true,
    Private: true,
    Layout: AppLayout
  },
  ComponentsLib: {
    title: 'Components Library',
    path: PathUrls.ComponentsLib,
    Component: componentsLib,
    Private: true,
    Layout: AppLayout
  },
  Analyse: {
    path: PathUrls.Analyse,
    title: 'Home',
    Component: CoreQuery,
    Private: true,
    Layout: AppLayout
  },
  Analyse1: {
    path: PathUrls.Analyse1,
    title: 'Home',
    Component: CoreQuery,
    Private: true,
    Layout: AppLayout
  },
  Analyse2: {
    path: PathUrls.Analyse2,
    title: 'Home',
    Component: CoreQuery,
    Private: true,
    Layout: AppLayout
  },
  Explain: {
    exact: true,
    path: PathUrls.Explain,
    title: 'Factors',
    Component: Factors,
    Private: true,
    Layout: AppLayout
  },
  ExplainInsightsV2: {
    exact: true,
    path: PathUrls.ExplainInsightsV2,
    title: 'ExplainV2',
    Component: FactorsInsightsNew,
    Private: true,
    Layout: AppLayout
  },
  ExplainInsights: {
    exact: true,
    path: '/explain/insights',
    title: 'Explain',
    Component: FactorsInsightsOld,
    Private: true,
    Layout: AppLayout
  },
  Welcome: {
    exact: true,
    path: '/welcome',
    Component: Welcome,
    Private: true,
    Layout: AppLayout
  },
  OnBoardFlow: {
    exact: true,
    path: '/welcome/visitoridentification/:step',
    Component: OnBoard,
    Private: true,
    Layout: AppLayout
  },
  Template: {
    exact: true,
    path: '/template',
    Component: DashboardTemplates,
    Private: true,
    Layout: AppLayout
  },
  SettingsGeneral: {
    exact: true,
    path: PathUrls.SettingsGeneral,
    Component: BasicSettings,
    Private: true,
    Layout: AppLayout
  },
  SettingsUser: {
    exact: true,
    path: PathUrls.SettingsUser,
    Component: UserSettings,
    Private: true,
    Layout: AppLayout
  },
  SettingsAttribution: {
    exact: true,
    path: PathUrls.SettingsAttribution,
    Component: AttributionSettings,
    Private: true,
    Layout: AppLayout
  },
  SettingsSdk: {
    exact: true,
    path: PathUrls.SettingsSdk,
    Component: SDKSettings,
    Private: true,
    Layout: AppLayout
  },
  SettingsIntegration: {
    exact: true,
    path: PathUrls.SettingsIntegration,
    Component: IntegrationSettings,
    Private: true,
    Layout: AppLayout
  },
  SettingsSharing: {
    exact: true,
    path: PathUrls.SettingsSharing,
    Component: Sharing,
    Private: true,
    Layout: AppLayout
  },
  SettingsInsights: {
    exact: true,
    path: PathUrls.SettingsInsights,
    name: 'dashboardSettings',
    Component: InsightsSettings,
    Private: true,
    Layout: AppLayout
  },
  ConfigureEvents: {
    exact: true,
    path: PathUrls.ConfigureEvents,
    Component: Events,
    Private: true,
    Layout: AppLayout
  },
  ConfigureProperties: {
    exact: true,
    path: PathUrls.ConfigureProperties,
    Component: PropertySettings,
    Private: true,
    Layout: AppLayout
  },
  ConfigureContentGroups: {
    exact: true,
    path: PathUrls.ConfigureContentGroups,
    Component: ContentGroups,
    Private: true,
    Layout: AppLayout
  },
  ConfigureTouchPoints: {
    exact: true,
    path: PathUrls.ConfigureTouchPoints,
    Component: Touchpoints,
    Private: true,
    Layout: AppLayout
  },
  ConfigureCustomKpi: {
    exact: true,
    path: PathUrls.ConfigureCustomKpi,
    Component: CustomKPI,
    Private: true,
    Layout: AppLayout
  },
  ConfigureDataPoints: {
    exact: true,
    path: PathUrls.ConfigureDataPoints,
    Component: ExplainDataPoints,
    Private: true,
    Layout: AppLayout
  },
  ConfigureAlerts: {
    exact: true,
    path: PathUrls.ConfigureAlerts,
    Component: Alerts,
    Private: true,
    Layout: AppLayout
  },
  ProfilePeople: {
    exact: true,
    path: PathUrls.ProfilePeople,
    Component: UserProfiles,
    Private: true,
    Layout: AppLayout
  },
  ProfileAccounts: {
    exact: true,
    path: PathUrls.ProfileAccounts,
    Component: AccountProfiles,
    Private: true,
    Layout: AppLayout
  },
  VisitorIdentificationReport: {
    exact: true,
    path: PathUrls.VisitorIdentificationReport,
    Layout: AppLayout,
    Private: false,
    Component: VisitorIdentificationReportComponent
  },
  //For backward compatibility for old url sent over mail
  SixSignalReportRedirection: {
    exact: true,
    path: '/reports/6_signal',
    Layout: AppLayout,
    Private: false,
    Component: SixSignalReportRedirection
  }
};

export const WhiteListedAccounts = [
  'baliga@factors.ai',
  'solutions@factors.ai',
  'sonali@factors.ai',
  'praveenr@factors.ai',
  'janani@factors.ai',
  'junaid@factors.ai'
];

export const TestEnvs = [
  'localhost',
  'factors-dev.com',
  'staging-app.factors.ai'
];

export const whiteListedProjects = [
  '1125899929000011',
  '2251799842000007',
  '2251799840000009',
  '12384898989000028',
  '2251799840000015',
  '2251799842000000',
  '12384898989000019',
  '1125899936000001',
  '2251799836000003',
  '1125899936000000',
  '12384898989000033',
  '12384898990000003',
  '12384898987000007',
  '2251799843000004',
  '1125899935000743',
  '2251799841000012',
  '12384898989000034',
  '2251799840000019',
  '12384898986000006'
];