import { combineReducers } from 'redux';
import GlobalReducer from './global';
import CoreQueryReducer from './coreQuery';
import agentActions from './agentActions';
import QueriesReducer from './queries';
import DashboardReducer from './dashboard';
import factors from './factors';
import events from './events';

const rootReducer = combineReducers({
  global: GlobalReducer,
  agent: agentActions,
  coreQuery: CoreQueryReducer,
  dashboard: DashboardReducer,
  queries: QueriesReducer,
  factors,
  events
});

export default rootReducer;
