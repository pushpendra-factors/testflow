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
import groups from './groups';

const rootReducer = combineReducers({
  global: GlobalReducer,
  agent: agentActions,
  coreQuery: CoreQueryReducer,
  analyticsQuery: AnalyticsReducer,
  dashboard: DashboardReducer,
  queries: QueriesReducer,
  settings: settingsReducer,
  factors,
  events,
  templates,
  insights,
  kpi,
  groups,
});

export default rootReducer;
