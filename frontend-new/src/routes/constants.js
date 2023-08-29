import React from 'react';
import lazyWithRetry from 'Utils/lazyWithRetry';

import DashboardTemplates from 'Views/DashboardTemplates/index';
import AttributionSettings from 'Views/Settings/ProjectSettings/AttributionSettings';
import BasicSettings from 'Views/Settings/ProjectSettings/BasicSettings';
import SDKSettings from 'Views/Settings/ProjectSettings/SDKSettings';
import UserSettings from 'Views/Settings/ProjectSettings/UserSettings';
import IntegrationSettings from 'Views/Settings/ProjectSettings/IntegrationSettings';
import Sharing from 'Views/Settings/ProjectSettings/Sharing';
import Events from 'Views/Settings/ProjectConfigure/Events';
import PropertySettings from 'Views/Settings/ProjectConfigure/PropertySettings';
import ContentGroups from 'Views/Settings/ProjectConfigure/ContentGroups';
import CustomKPI from 'Views/Settings/ProjectConfigure/CustomKPI';
import Alerts from 'Views/Settings/ProjectSettings/Alerts';
import ExplainDataPoints from 'Views/Settings/ProjectConfigure/ExplainDataPoints';
import UserProfiles from 'Components/Profile/UserProfiles';
import AccountProfiles from 'Components/Profile/AccountProfiles';
import Touchpoints from 'Views/Settings/ProjectConfigure/Touchpoints';
import AppLayout from 'Views/AppLayout';
import { PathUrls } from './pathUrls';
import AccountDetails from 'Components/Profile/AccountProfiles/AccountDetails';
import ContactDetails from 'Components/Profile/UserProfiles/ContactDetails';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import LockedStateComponent from 'Components/GenericComponents/LockedStateVideoComponent';
import PricingComponent from 'Views/Settings/ProjectSettings/Pricing';
import EngagementConfig from 'Views/Settings/ProjectConfigure/Engagement';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';
import Onboarding from '../features/onboarding/ui';

//locked screen images
import LockedExplainImage from '../assets/images/locked_explain.png';
import LockedPathAnalysisImage from '../assets/images/locked_path_analysis.png';

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

const FactorsInsightsNew = lazyWithRetry(() =>
  import('../Views/Factors/FactorsInsightsNew')
);

const PathAnalysis = lazyWithRetry(() => import('../Views/PathAnalysis'));
const FeatureLockedPathAnalysis = withFeatureLockHOC(PathAnalysis, {
  featureName: FEATURES.FEATURE_PATH_ANALYSIS,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Path Analysis'}
      embeddedLink={LockedPathAnalysisImage}
      description='Gain valuable insights into customer journeys and optimize conversion paths. Understand how prospects navigate your website, attribute revenue to specific marketing efforts, optimize content and campaigns, and deliver personalized experiences for increased conversions and marketing success'
    />
  )
});
const PathAnalysisReport = lazyWithRetry(() =>
  import('../Views/PathAnalysis/PathAnalysisReport')
);
const FeatureLockedPathAnalysisReport = withFeatureLockHOC(PathAnalysisReport, {
  featureName: FEATURES.FEATURE_PATH_ANALYSIS,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Path Analysis'}
      embeddedLink={LockedPathAnalysisImage}
      description='Gain valuable insights into customer journeys and optimize conversion paths. Understand how prospects navigate your website, attribute revenue to specific marketing efforts, optimize content and campaigns, and deliver personalized experiences for increased conversions and marketing success'
    />
  )
});
const FeatureLockedPropertySettings = withFeatureLockHOC(PropertySettings, {
  featureName: FEATURES.CONF_CUSTOM_PROPERTIES,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Properties'
      description='Harness the full potential of your advertising data with Custom Properties. By associating distinct attributes with your data, you gain precise control over configuring and analyzing your ad campaigns. Customize and tailor your data to align perfectly with your business objectives, ensuring optimal insights and enhanced advertising optimization.'
      learnMoreLink='https://help.factors.ai/en/articles/7284109-custom-properties'
    />
  )
});

const FeatureLockedConfigureContentGroups = withFeatureLockHOC(ContentGroups, {
  featureName: FEATURES.CONF_CONTENT_GROUPS,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Content Groups'
      description='Create logical collections of related URLs, such as blog articles or product pages, to analyze their impact on leads, revenue, and pipeline stages. Compare the performance of different content groups, identify optimization opportunities, and enhance your content marketing efforts to drive better results.'
      learnMoreLink='https://help.factors.ai/en/articles/7284125-content-groups'
    />
  )
});

const FeatureLockedConfigureTouchpoints = withFeatureLockHOC(Touchpoints, {
  featureName: FEATURES.CONF_TOUCHPOINTS,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Touchpoints'
      description='Effortlessly map and standardize your marketing parameters. Connect and align UTMs and other parameters used across your marketing efforts to a standardized set. Query and filter by different parameter values within Factors, enabling seamless tracking and analysis of customer touchpoints'
    />
  )
});

const FeatureLockedConfigureCustomKPI = withFeatureLockHOC(CustomKPI, {
  featureName: FEATURES.CONF_CUSTOM_KPIPS,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Custom KPIs'
      description="Create personalized metrics tailored to your specific objectives, whether it's conversion rates, engagement metrics, or revenue targets. Monitor progress, measure success, and gain actionable insights to drive continuous improvement and achieve your business milestones."
      learnMoreLink='https://help.factors.ai/en/articles/7284181-custom-kpis'
    />
  )
});

const FeatureLockedConfigureEvents = withFeatureLockHOC(Events, {
  featureName: FEATURES.CONF_CUSTOM_EVENTS,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Events'
      description='Track and analyze user interactions in a way that aligns perfectly with your business objectives. Define and capture custom events that matter most to your business, such as clicks, form submissions, lifecycle stage changes, or other specific actions.'
      learnMoreLink='https://help.factors.ai/en/articles/7284092-custom-events'
    />
  )
});

const FeatureLockedConfigureExplainDataPoints = withFeatureLockHOC(
  ExplainDataPoints,
  {
    featureName: FEATURES.CONF_TOUCHPOINTS,
    LockedComponent: () => (
      <CommonLockedComponent
        title='Top Events and Properties'
        description='Elevate the importance of key events and properties in your project with our Top Events and Properties feature. By designating specific events and properties as top priorities, you can ensure they are closely monitored and tracked. These vital metrics will be prominently displayed in the Explain section of Factors, providing you with instant visibility and easy access to the most critical data points.'
        learnMoreLink='https://help.factors.ai/en/articles/6294993-top-events-and-properties'
      />
    )
  }
);

const FeatureLockConfigurationAttribution = withFeatureLockHOC(
  AttributionSettings,
  {
    featureName: FEATURES.CONF_ATTRUBUTION_SETTINGS,
    LockedComponent: () => (
      <CommonLockedComponent
        title='Attribution'
        description='Attribute revenue and conversions to the right marketing channels, campaigns, and touchpoints to gain a clear understanding of what drives success. Identify the most effective marketing strategies, optimize your budget allocation, and make data-driven decisions to maximize ROI and achieve your business goals.'
      />
    )
  }
);

const FeatureLockedReportSharing = withFeatureLockHOC(Sharing, {
  featureName: FEATURES.FEATURE_REPORT_SHARING,
  LockedComponent: () => <CommonLockedComponent title='Sharing' />
});

const FeatureLockConfigurationAlerts = withFeatureLockHOC(Alerts, {
  featureName: FEATURES.CONF_ALERTS,
  LockedComponent: () => (
    <CommonLockedComponent
      title='Alerts'
      description='With real-time alerts in Slack, stay informed the moment a prospect visits a high-intent page on your website or when a significant change occurs in a KPI that matters to your organization. Be instantly notified, take immediate action, and seize every opportunity to drive conversions, optimize performance, and achieve your business objectives.'
      learnMoreLink='https://help.factors.ai/en/articles/7284705-alerts'
    />
  )
});

const FeatureLockedConfigurationEngagement = withFeatureLockHOC(
  EngagementConfig,
  {
    featureName: FEATURES.CONF_TOUCHPOINTS,
    LockedComponent: () => (
      <CommonLockedComponent
        title='Engagement Scoring'
        description='Some events matter more than others, and are better indicators of buying intent. Configure scores for them, tag them as intent signals, and more.'
      />
    )
  }
);

const FeatureLockedFactorsInsightsNew = withFeatureLockHOC(FactorsInsightsNew, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Explain'}
      embeddedLink={LockedExplainImage}
      description='All your important metrics at a glance. The dashboard is where you save your analyses for quick and easy viewing. Create multiple dashboards for different needs, and toggle through them as you wish. Making the right decisions just became easier.'
    />
  )
});
const FactorsInsightsOld = lazyWithRetry(() =>
  import('../Views/Factors/FactorsInsightsOld')
);
const FeatureLockedFactorsInsightsOld = withFeatureLockHOC(FactorsInsightsOld, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Explain'}
      embeddedLink={LockedExplainImage}
      description='All your important metrics at a glance. The dashboard is where you save your analyses for quick and easy viewing. Create multiple dashboards for different needs, and toggle through them as you wish. Making the right decisions just became easier.'
    />
  )
});
const CoreQuery = lazyWithRetry(() => import('../Views/CoreQuery'));
const Dashboard = lazyWithRetry(() => import('../Views/Dashboard'));
const Factors = lazyWithRetry(() => import('../Views/Factors'));
const FeatureLockedFactors = withFeatureLockHOC(Factors, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Explain'}
      embeddedLink={LockedExplainImage}
      description='All your important metrics at a glance. The dashboard is where you save your analyses for quick and easy viewing. Create multiple dashboards for different needs, and toggle through them as you wish. Making the right decisions just became easier.'
    />
  )
});
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
    Component: FeatureLockedFactors,
    Private: true,
    Layout: AppLayout
  },
  ExplainInsightsV2: {
    exact: true,
    path: PathUrls.ExplainInsightsV2,
    title: 'ExplainV2',
    Component: FeatureLockedFactorsInsightsNew,
    Private: true,
    Layout: AppLayout
  },
  ExplainInsights: {
    exact: true,
    path: '/explain/insights',
    title: 'Explain',
    Component: FeatureLockedFactorsInsightsOld,
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
    Component: FeatureLockedReportSharing,
    Private: true,
    Layout: AppLayout
  },
  SettingsPricing: {
    exact: true,
    path: PathUrls.SettingsPricing,
    name: 'pricingSettings',
    Component: PricingComponent,
    Private: true,
    Layout: AppLayout
  },
  ConfigureEvents: {
    exact: true,
    path: PathUrls.ConfigureEvents,
    Component: FeatureLockedConfigureEvents,
    Private: true,
    Layout: AppLayout
  },
  ConfigureProperties: {
    exact: true,
    path: PathUrls.ConfigureProperties,
    Component: FeatureLockedPropertySettings,
    Private: true,
    Layout: AppLayout
  },
  ConfigureContentGroups: {
    exact: true,
    path: PathUrls.ConfigureContentGroups,
    Component: FeatureLockedConfigureContentGroups,
    Private: true,
    Layout: AppLayout
  },
  ConfigureTouchPoints: {
    exact: true,
    path: PathUrls.ConfigureTouchPoints,
    Component: FeatureLockedConfigureTouchpoints,
    Private: true,
    Layout: AppLayout
  },
  ConfigureCustomKpi: {
    exact: true,
    path: PathUrls.ConfigureCustomKpi,
    Component: FeatureLockedConfigureCustomKPI,
    Private: true,
    Layout: AppLayout
  },
  ConfigureDataPoints: {
    exact: true,
    path: PathUrls.ConfigureDataPoints,
    Component: FeatureLockedConfigureExplainDataPoints,
    Private: true,
    Layout: AppLayout
  },
  ConfigureAlerts: {
    exact: true,
    path: PathUrls.ConfigureAlerts,
    Component: FeatureLockConfigurationAlerts,
    Private: true,
    Layout: AppLayout
  },
  ConfigureEngagements: {
    exact: true,
    path: PathUrls.ConfigureEngagements,
    Component: FeatureLockedConfigurationEngagement,
    Private: true,
    Layout: AppLayout
  },
  ConfigurationAttribution: {
    exact: true,
    path: PathUrls.ConfigureAttribution,
    Component: FeatureLockConfigurationAttribution,
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
  ProfileUserDetails: {
    path: '/profiles/people/:id',
    Component: ContactDetails,
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
  ProfileAccountsDetails: {
    path: '/profiles/accounts/:id',
    Component: AccountDetails,
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
  PathAnalysis: {
    exact: true,
    path: PathUrls.PathAnalysis,
    Layout: AppLayout,
    Private: true,
    Component: FeatureLockedPathAnalysis
  },
  PathAnalysisInsights: {
    exact: true,
    path: PathUrls.PathAnalysisInsights,
    Layout: AppLayout,
    Private: true,
    Component: FeatureLockedPathAnalysisReport
  },
  Onboarding: {
    exact: true,
    path: PathUrls.Onboarding,
    Layout: AppLayout,
    Private: true,
    Component: Onboarding
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

export const SolutionsAccountId = 'solutions@factors.ai';

export const WhiteListedAccounts = [
  'baliga@factors.ai',
  SolutionsAccountId,
  'sonali@factors.ai',
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
