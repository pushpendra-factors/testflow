import { combineReducers } from 'redux';
import GlobalReducer from './global';

const rootReducer = combineReducers({
  global: GlobalReducer
});

export default rootReducer;
