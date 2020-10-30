import { combineReducers } from 'redux';
import GlobalReducer from './global';
import CoreQueryReducer from './coreQuery';
import agentActions from './agentActions';
import QueriesReducer from './queries';

const rootReducer = combineReducers({
  global: GlobalReducer,
  agent: agentActions,
  coreQuery: CoreQueryReducer,
  queries: QueriesReducer
});

export default rootReducer;
