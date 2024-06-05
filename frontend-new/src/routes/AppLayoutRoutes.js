import React, { useEffect } from 'react';
import lazyWithRetry from 'Utils/lazyWithRetry';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import LockedStateComponent from 'Components/GenericComponents/LockedStateVideoComponent';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';

import { Switch, Redirect, Route } from 'react-router-dom';
import PrivateRoute from 'Components/PrivateRoute';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { useDispatch } from 'react-redux';
import { UPDATE_ALL_ROUTES } from 'Reducers/types';
import { PathUrls } from './pathUrls';
import LockedPathAnalysisImage from '../assets/images/locked_path_analysis.png';
import LockedExplainImage from '../assets/images/locked_explain.png';
import { AdminLock, featureLock } from './feature';
import LockedAttributionImage from '../assets/images/locked_attribution.png';
import { renderRoutes } from './utils';

// Profile-Account
const AccountDetails = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-account" */ '../components/Profile/AccountProfiles/AccountDetails'
    )
);
const AccountProfiles = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-account" */ '../components/Profile/AccountProfiles'
    )
);

// Profile-People
const ContactDetails = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-people" */ '../components/Profile/UserProfiles/ContactDetails'
    )
);
const UserProfiles = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-people" */ '../components/Profile/UserProfiles'
    )
);
const VisitorIdentificationReportComponent = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-people" */ '../features/6signal-report/ui'
    )
);
const SixSignalReportRedirection = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "profile-people" */ '../features/6signal-report/ui/SixSignalRedirection'
    )
);

// Dashboard
const Dashboard = lazyWithRetry(
  () => import(/* webpackChunkName: "dashboard" */ '../Views/Dashboard')
);
const PreBuildDashboardReport = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "dashboard" */ '../Views/PreBuildDashboard/ui/Report'
    )
);
const CoreQueryNew = lazyWithRetry(
  () => import(/* webpackChunkName: "dashboard" */ '../features/analyse')
);
const CoreQuery = lazyWithRetry(
  () => import(/* webpackChunkName: "dashboard" */ '../Views/CoreQuery')
);
const PreBuildDashboard = lazyWithRetry(
  () =>
    import(/* webpackChunkName: "dashboard" */ '../Views/PreBuildDashboard/ui')
);

// Path Analysis
const PathAnalysis = lazyWithRetry(
  () => import(/* webpackChunkName: "path-analysis" */ '../Views/PathAnalysis')
);
const PathAnalysisReport = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "path-analysis" */ '../Views/PathAnalysis/PathAnalysisReport'
    )
);

// Explain
const FactorsInsightsNew = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "explain" */ '../Views/Factors/FactorsInsightsNew'
    )
);
const Factors = lazyWithRetry(
  () => import(/* webpackChunkName: "explain" */ '../Views/Factors')
);
const FactorsInsightsOld = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "explain" */ '../Views/Factors/FactorsInsightsOld'
    )
);

// Attribution
const Attribution = lazyWithRetry(
  () =>
    import(/* webpackChunkName: "Attribution" */ '../features/attribution/ui')
);

// Alerts
const Alerts = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "Alerts" */ '../Views/Settings/ProjectSettings/Alerts'
    )
);

// Workflows
const Workflows = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "Workflows" */ '../Views/Settings/ProjectSettings/Workflows'
    )
);

// Settings
const BasicSettings = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings" */ '../Views/Settings/ProjectSettings/BasicSettings'
    )
);
const UserSettings = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings" */ '../Views/Settings/ProjectSettings/UserSettings'
    )
);

const Sharing = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings" */ '../Views/Settings/ProjectSettings/Sharing'
    )
);
const PricingComponent = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings" */ '../Views/Settings/ProjectSettings/Pricing'
    )
);

// Integration Screen
const IntegrationSettings = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "integration" */ '../Views/Settings/ProjectSettings/IntegrationSettings/integrationRoute'
    )
);

const IntegrationRedirection = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "integration" */ '../Views/Settings/ProjectSettings/IntegrationSettings/IntegrationCallbackRedirection'
    )
);

// settings personal

const SettingsPersonalUser = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings-personal" */ '../Views/Settings/UserSettings/ViewUserDetails'
    )
);

const SettingsPersonalProjects = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings-personal" */ '../Views/Settings/UserSettings/ProjectDetails'
    )
);

// Settings data management

const AttributionSettings = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings-attribution" */ '../Views/Settings/ProjectSettings/AttributionSettings'
    )
);
const ConfigurePlans = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "configuration-plan" */ '../Views/Settings/ProjectSettings/ConfigurePlans'
    )
);
const ConfigurePlanAdmin = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "configuration-plan" */
      '../Views/Settings/ProjectSettings/ConfigurePlans/ConfigurePlanAdmin'
    )
);

const Touchpoints = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings-touchpoint" */ '../Views/Settings/ProjectConfigure/Touchpoints'
    )
);

const CustomDefinition = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "settings-custom-definition" */ '../Views/Settings/ProjectConfigure/CustomDefinition'
    )
);

// Shared-Components
const componentsLib = lazyWithRetry(
  () =>
    import(/* webpackChunkName: "shared-component" */ '../Views/componentsLib')
);

const Checklist = lazyWithRetry(
  () =>
    import(/* webpackChunkName: "shared-component" */ '../features/Checklist')
);

const DashboardTemplates = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "shared-component" */ '../Views/DashboardTemplates/index'
    )
);

// Paragon-workflow
const WorkflowParagon = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "paragon-workflow" */ '../Views/Pages/WorkflowParagon'
    )
);

// Onboarding
const Onboarding = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "Onboarding" */ '../features/onboarding/ui/index'
    )
);

const FeatureLockedAttributionComponent = withFeatureLockHOC(Attribution, {
  featureName: FEATURES.FEATURE_ATTRIBUTION,
  LockedComponent: () => (
    <LockedStateComponent
      title='Attribution'
      description='Attribute revenue and conversions to the right marketing channels, campaigns, and touchpoints to gain a clear understanding of what drives success. Identify the most effective marketing strategies, optimize your budget allocation, and make data-driven decisions to maximize ROI and achieve your business goals.'
      embeddedLink={LockedAttributionImage}
    />
  )
});

const FeatureLockedPathAnalysis = withFeatureLockHOC(PathAnalysis, {
  featureName: FEATURES.FEATURE_PATH_ANALYSIS,
  LockedComponent: (props) => (
    <LockedStateComponent
      title='Path Analysis'
      embeddedLink={LockedPathAnalysisImage}
      description='Gain valuable insights into customer journeys and optimize conversion paths. Understand how prospects navigate your website, attribute revenue to specific marketing efforts, optimize content and campaigns, and deliver personalized experiences for increased conversions and marketing success'
      {...props}
    />
  )
});

const FeatureLockedPathAnalysisReport = withFeatureLockHOC(PathAnalysisReport, {
  featureName: FEATURES.FEATURE_PATH_ANALYSIS,
  LockedComponent: (props) => (
    <LockedStateComponent
      title='Path Analysis'
      embeddedLink={LockedPathAnalysisImage}
      description='Gain valuable insights into customer journeys and optimize conversion paths. Understand how prospects navigate your website, attribute revenue to specific marketing efforts, optimize content and campaigns, and deliver personalized experiences for increased conversions and marketing success'
      {...props}
    />
  )
});

const FeatureLockConfigurationAttribution = withFeatureLockHOC(
  AttributionSettings,
  {
    featureName: FEATURES.FEATURE_ATTRIBUTION,
    LockedComponent: (props) => (
      <CommonLockedComponent
        title='Attribution'
        description='Attribute revenue and conversions to the right marketing channels, campaigns, and touchpoints to gain a clear understanding of what drives success. Identify the most effective marketing strategies, optimize your budget allocation, and make data-driven decisions to maximize ROI and achieve your business goals.'
        {...props}
      />
    )
  }
);

const FeatureLockedReportSharing = withFeatureLockHOC(Sharing, {
  featureName: FEATURES.FEATURE_REPORT_SHARING,
  LockedComponent: (props) => (
    <CommonLockedComponent title='Sharing' {...props} />
  )
});

const FeatureLockConfigurationAlerts = withFeatureLockHOC(Alerts, {
  featureName: FEATURES.FEATURE_EVENT_BASED_ALERTS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      title='Alerts'
      description='With real-time alerts in Slack, stay informed the moment a prospect visits a high-intent page on your website or when a significant change occurs in a KPI that matters to your organization. Be instantly notified, take immediate action, and seize every opportunity to drive conversions, optimize performance, and achieve your business objectives.'
      learnMoreLink='https://help.factors.ai/en/articles/7284705-alerts'
      {...props}
    />
  )
});
const FeatureLockConfigurationWorkflows = withFeatureLockHOC(Workflows, {
  featureName: FEATURES.FEATURE_WORKFLOWS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      title='Workflows'
      description='With real-time alerts in Slack, stay informed the moment a prospect visits a high-intent page on your website or when a significant change occurs in a KPI that matters to your organization. Be instantly notified, take immediate action, and seize every opportunity to drive conversions, optimize performance, and achieve your business objectives.'
      learnMoreLink='https://help.factors.ai/en/articles/7284705-alerts'
      {...props}
    />
  )
});

const FeatureLockedFactorsInsightsNew = withFeatureLockHOC(FactorsInsightsNew, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: (props) => (
    <LockedStateComponent
      title='Explain'
      embeddedLink={LockedExplainImage}
      description='Uncover the driving factors behind your conversion goals with Explain. Gain deep insights into the elements that contribute to your objectives, empowering you to make informed decisions and optimize your strategies for success.'
      {...props}
    />
  )
});

const FeatureLockedFactorsInsightsOld = withFeatureLockHOC(FactorsInsightsOld, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: (props) => (
    <LockedStateComponent
      title='Explain'
      embeddedLink={LockedExplainImage}
      description='Uncover the driving factors behind your conversion goals with Explain. Gain deep insights into the elements that contribute to your objectives, empowering you to make informed decisions and optimize your strategies for success.'
      {...props}
    />
  )
});

const FeatureLockedFactors = withFeatureLockHOC(Factors, {
  featureName: FEATURES.FEATURE_EXPLAIN,
  LockedComponent: (props) => (
    <LockedStateComponent
      title='Explain'
      embeddedLink={LockedExplainImage}
      description='Uncover the driving factors behind your conversion goals with Explain. Gain deep insights into the elements that contribute to your objectives, empowering you to make informed decisions and optimize your strategies for success.'
      {...props}
    />
  )
});

const FeatureLockedPreBuildDashboard = withFeatureLockHOC(PreBuildDashboard, {
  featureName: FEATURES.FEATURE_WEB_ANALYTICS_DASHBOARD,
  LockedComponent: (props) => (
    <CommonLockedComponent
      title='Traffic Dashboard'
      description='This dashboard tracks a few commonly accessed metrics. The widgets you see are either Event, Funnel, or Attribution reports. They were built by a user and saved into this space.'
      {...props}
    />
  )
});

export const APP_LAYOUT_ROUTES = {
  Workflow: {
    exact: true,
    path: '/workflow',
    Private: true,
    Component: WorkflowParagon
  },
  // moved this to top for matching before /reports/:dashboard_id
  VisitorIdentificationReport: {
    exact: true,
    path: PathUrls.VisitorIdentificationReport,
    Private: false,
    Component: VisitorIdentificationReportComponent
  },
  PreBuildDashboard: {
    title: 'Dashboard',
    path: PathUrls.PreBuildDashboard,
    Component: FeatureLockedPreBuildDashboard,
    exact: true,
    Private: true
  },
  PreBuildDashboardReport: {
    exact: true,
    path: PathUrls.PreBuildDashboardReport,
    Private: true,
    Component: PreBuildDashboardReport
  },
  Dashboard: {
    title: 'Dashboard',
    path: PathUrls.Dashboard,
    Component: Dashboard,
    Private: true
  },
  DashboardUrl: {
    title: 'Dashboard',
    path: PathUrls.DashboardURL,
    Component: Dashboard,
    Private: true
  },
  ComponentsLib: {
    title: 'Components Library',
    path: PathUrls.ComponentsLib,
    Component: componentsLib,
    Private: true
  },
  Analyse: {
    path: PathUrls.Analyse,
    title: 'Home',
    Component: CoreQueryNew,
    Private: true
  },
  Analyse1: {
    path: PathUrls.Analyse1,
    title: 'Home',
    Component: CoreQuery,
    Private: true
  },
  Analyse2: {
    path: PathUrls.Analyse2,
    title: 'Home',
    Component: CoreQuery,
    Private: true
  },
  Explain: {
    exact: true,
    path: PathUrls.Explain,
    title: 'Factors',
    Component: FeatureLockedFactors,
    Private: true
  },
  ExplainInsightsV2: {
    exact: true,
    path: PathUrls.ExplainInsightsV2,
    title: 'ExplainV2',
    Component: FeatureLockedFactorsInsightsNew,
    Private: true
  },
  ExplainInsights: {
    exact: true,
    path: '/explain/insights',
    title: 'Explain',
    Component: FeatureLockedFactorsInsightsOld,
    Private: true
  },
  Template: {
    exact: true,
    path: '/template',
    Component: DashboardTemplates,
    Private: true
  },
  SettingsGeneral: {
    exact: true,
    path: PathUrls.SettingsGeneral,
    Component: BasicSettings,
    Private: true
  },
  SettingsMembers: {
    exact: true,
    path: PathUrls.SettingsMembers,
    Component: UserSettings,
    Private: true
  },
  SettingsIntegration: {
    path: PathUrls.SettingsIntegration,
    Component: IntegrationSettings,
    Private: true
  },
  IntegrationRedirection: {
    exact: true,
    path: PathUrls.IntegrationCallbackRedirection,
    Component: IntegrationRedirection,
    Private: true
  },
  SettingsSharing: {
    exact: true,
    path: PathUrls.SettingsSharing,
    Component: FeatureLockedReportSharing,
    Private: true
  },
  SettingsPricing: {
    exact: true,
    path: PathUrls.SettingsPricing,
    name: 'pricingSettings',
    Component: PricingComponent,
    Private: true
  },

  // personal settings
  SettingsPersonalUser: {
    exact: true,
    path: PathUrls.SettingsPersonalUser,
    name: 'settingsPersonalUser',
    Component: SettingsPersonalUser,
    Private: true
  },
  SettingsPersonalProjects: {
    exact: true,
    path: PathUrls.SettingsPersonalProjects,
    name: 'settingsPersonalUser',
    Component: SettingsPersonalProjects,
    Private: true
  },

  SettingsTouchpointDefinition: {
    exact: true,
    path: PathUrls.SettingsTouchpointDefinition,
    Component: Touchpoints,
    Private: true
  },
  SettingsCustomDefinition: {
    exact: true,
    path: PathUrls.SettingsCustomDefinition,
    Component: CustomDefinition,
    Private: true
  },

  Alerts: {
    exact: true,
    path: PathUrls.Alerts,
    Component: FeatureLockConfigurationAlerts,
    Private: true
  },
  Workflows: {
    exact: true,
    path: PathUrls.Workflows,
    Component: FeatureLockConfigurationWorkflows,
    Private: false
  },

  SettingsAttribution: {
    exact: true,
    path: PathUrls.SettingsAttribution,
    Component: FeatureLockConfigurationAttribution,
    Private: true
  },
  ProfilePeople: {
    exact: true,
    path: PathUrls.ProfilePeople,
    Component: UserProfiles,
    Private: true
  },
  ProfileUserDetails: {
    path: '/profiles/people/:id',
    Component: ContactDetails,
    Private: true
  },
  ProfileAccounts: {
    exact: true,
    path: PathUrls.ProfileAccounts,
    Component: AccountProfiles,
    Private: true
  },
  ProfileAccountsSegmentsURL: {
    title: 'Accounts',
    exact: true,
    path: PathUrls.ProfileAccountsSegmentsURL,
    Component: AccountProfiles,
    Private: true
  },
  ProfileAccountsDetails: {
    path: PathUrls.ProfileAccountDetailsURL,
    Component: AccountDetails,
    Private: true
  },

  PathAnalysis: {
    exact: true,
    path: PathUrls.PathAnalysis,
    Private: true,
    Component: FeatureLockedPathAnalysis
  },
  PathAnalysisInsights: {
    exact: true,
    path: PathUrls.PathAnalysisInsights,
    Private: true,
    Component: FeatureLockedPathAnalysisReport
  },
  Onboarding: {
    exact: true,
    path: PathUrls.Onboarding,
    Private: true,
    Component: Onboarding
  },
  ConfigurePlanAdmin: {
    exact: true,
    path: PathUrls.ConfigurePlansAdmin,
    Private: true,
    Component: ConfigurePlanAdmin
  },
  // For backward compatibility for old url sent over mail
  SixSignalReportRedirection: {
    exact: true,
    path: '/reports/6_signal',
    Private: false,
    Component: SixSignalReportRedirection
  }
};

export function AppLayoutRoutes({
  activeAgent,
  active_project,
  currentProjectSettings
}) {
  const dispatch = useDispatch();

  useEffect(() => {
    if (featureLock(activeAgent)) {
      const allRoutes = [];
      allRoutes.push(ATTRIBUTION_ROUTES.base);

      dispatch({ type: UPDATE_ALL_ROUTES, payload: allRoutes });
    }
  }, [activeAgent]);
  useEffect(() => {
    const allRoutes = [];
    Object.keys(APP_LAYOUT_ROUTES).forEach((key) => {
      if (APP_LAYOUT_ROUTES[key]?.path) {
        allRoutes.push(APP_LAYOUT_ROUTES[key]?.path);
      }
    });

    dispatch({ type: UPDATE_ALL_ROUTES, payload: allRoutes });
  }, []);
  return (
    <Switch>
      {renderRoutes(APP_LAYOUT_ROUTES)}
      {/* Additional Conditional routes  */}

      <PrivateRoute
        path={ATTRIBUTION_ROUTES.base}
        name='attribution'
        component={FeatureLockedAttributionComponent}
      />

      {AdminLock(activeAgent) ? (
        <PrivateRoute
          path={PathUrls.ConfigurePlans}
          name='Configure Plans'
          component={ConfigurePlans}
        />
      ) : null}

      <PrivateRoute path={PathUrls.Checklist} component={Checklist} />

      {/* if no route match, redirect to home-screen */}
      <Route render={() => <Redirect to='/' />} />
    </Switch>
  );
}
