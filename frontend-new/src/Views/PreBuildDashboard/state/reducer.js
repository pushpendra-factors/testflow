import {
  ACTIVE_PRE_DASHBOARD_CHANGE,
  SET_ACTIVE_PROJECT
} from 'Reducers/types';
import { getRearrangedData } from 'Reducers/dashboard/utils';
import { EMPTY_ARRAY, EMPTY_OBJECT } from 'Utils/global';
import {
  DASHBOARD_CONFIG_LOADED,
  DASHBOARD_CONFIG_LOADING,
  DASHBOARD_CONFIG_LOADING_FAILED,
  SET_FILTER_PAYLOAD,
  SET_REPORT_FILTER_PAYLOAD
} from './services';

export const defaultState = {
  config: {
    loading: false,
    error: false,
    data: EMPTY_OBJECT
  },
  widget: EMPTY_ARRAY,
  activePreBuildDashboard: EMPTY_OBJECT,
  filters: [],
  reportFilters: []
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case ACTIVE_PRE_DASHBOARD_CHANGE:
      return {
        ...state,
        activePreBuildDashboard: action.payload
      };
    case DASHBOARD_CONFIG_LOADING:
      return {
        ...state,
        config: {
          ...defaultState.config,
          loading: true
        }
      };
    case DASHBOARD_CONFIG_LOADING_FAILED:
      return {
        ...state,
        config: {
          ...defaultState.config,
          error: true
        }
      };
    case DASHBOARD_CONFIG_LOADED:
      return {
        ...state,
        config: {
          ...defaultState.data,
          data: action.payload,
          loading: false,
          error: false
        },
        widget: getRearrangedData(
          action.payload.result.wid,
          state.activePreBuildDashboard
        )
      };
    case SET_FILTER_PAYLOAD:
      return {
        ...state,
        filters: action.payload
      };
    case SET_REPORT_FILTER_PAYLOAD:
      return {
        ...state,
        reportFilters: action.payload
      };
    case SET_ACTIVE_PROJECT: {
      return {
        ...defaultState
      };
    }
    default:
      return state;
  }
}
