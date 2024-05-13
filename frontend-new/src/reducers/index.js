import { combineReducers } from 'redux';
import GlobalReducer from './global';
import CoreQueryReducer from './coreQuery';
import AnalyticsReducer from './analyticsQuery';
import agentActions from './agentActions';
import QueriesReducer from './queries';
import DashboardReducer from './dashboard';
import factors from './factors';
import events from './events';
import settingsReducer from './settings';
import templates from './templates';
import insights from './insights';
import kpi from './kpi';
import pathAnalysis from './pathAnalysis';
import groups from './groups';
import timelines from './timelines';
import dashboardTemplateReducer from './dashboard_templates';
import dashboardTemplatesControllerReducer from './dashboard_templates_modal';
import attributionReducer from '../features/attribution/state/reducer';
import globalSearch from './globalSearch';
import allRoutes from './allRoutes';
import accountProfilesViewReducer from './accountProfilesView';
import userProfilesViewReducer from './userProfilesView';
import FeatureConfigReducer from './featureConfig';
import { USER_LOGOUT } from './types';
import preBuildDashboardConfig from '../Views/PreBuildDashboard/state/reducer';
import PlansConfigReducer from './plansConfig';
import alertTemplates from './alertTemplates';
import workflows from './workflows';

const appReducer = combineReducers({
  global: GlobalReducer,
  agent: agentActions,
  coreQuery: CoreQueryReducer,
  analyticsQuery: AnalyticsReducer,
  dashboard: DashboardReducer,
  preBuildDashboardConfig,
  queries: QueriesReducer,
  settings: settingsReducer,
  factors,
  events,
  templates,
  insights,
  kpi,
  groups,
  timelines,
  dashboardTemplates: dashboardTemplateReducer,
  dashboardTemplatesController: dashboardTemplatesControllerReducer,
  pathAnalysis,
  attributionDashboard: attributionReducer,
  globalSearch,
  allRoutes,
  accountProfilesView: accountProfilesViewReducer,
  userProfilesView: userProfilesViewReducer,
  featureConfig: FeatureConfigReducer,
  plansConfig: PlansConfigReducer,
  alertTemplates: alertTemplates,
  workflows: workflows
});

const rootReducer = (state, action) => {
  if (action.type === USER_LOGOUT) {
    // for all keys defined in your persistConfig(s)
    localStorage.removeItem('persist:root');
    // storage.removeItem('persist:otherKey')

    return appReducer(undefined, action);
  }
  return appReducer(state, action);
};

export default rootReducer;
