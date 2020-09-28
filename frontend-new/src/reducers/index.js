import { combineReducers } from 'redux';
import GlobalReducer from './global';
import CoreQueryReducer from './coreQuery';

const rootReducer = combineReducers({
  global: GlobalReducer,
  coreQuery: CoreQueryReducer
});

export default rootReducer;
