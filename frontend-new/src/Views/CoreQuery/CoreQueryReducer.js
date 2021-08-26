import {
  SET_ATTRIBUTION_METRICS,
  SET_NAVIGATED_FROM_DASHBOARD,
  SET_COMPARISON_ENABLED,
  COMPARISON_DATA_LOADING,
  COMPARISON_DATA_FETCHED,
  RESET_COMPARISON_DATA,
  SET_COMPARISON_SUPPORTED,
  SET_COMPARE_DURATION,
  UPDATE_CHART_TYPES,
  SET_SAVED_QUERY_SETTINGS,
} from './constants';

export default function (state, action) {
  const { payload } = action;
  switch (action.type) {
    case SET_ATTRIBUTION_METRICS:
      return {
        ...state,
        attributionMetrics: payload,
      };
    case SET_NAVIGATED_FROM_DASHBOARD:
      return {
        ...state,
        navigatedFromDashboard: payload,
      };
    case SET_COMPARISON_SUPPORTED:
      return {
        ...state,
        comparison_supported: payload,
      };
    case SET_COMPARISON_ENABLED:
      return {
        ...state,
        comparison_enabled: payload,
      };
    case COMPARISON_DATA_LOADING:
      return {
        ...state,
        comparison_data: {
          loading: true,
          error: false,
          data: null,
        },
      };
    case COMPARISON_DATA_FETCHED:
      return {
        ...state,
        comparison_data: {
          loading: false,
          error: false,
          data: payload,
        },
      };
    case RESET_COMPARISON_DATA:
      return {
        ...state,
        comparison_duration: null,
        comparison_enabled: false,
        comparison_data: {
          loading: false,
          error: false,
          data: null,
        },
      };
    case SET_COMPARE_DURATION:
      return {
        ...state,
        comparison_duration: payload,
      };
    case UPDATE_CHART_TYPES:
      return {
        ...state,
        chartTypes: payload,
      };
    case SET_SAVED_QUERY_SETTINGS: {
      return {
        ...state,
        savedQuerySettings: payload,
      };
    }
    default:
      return state;
  }
}
