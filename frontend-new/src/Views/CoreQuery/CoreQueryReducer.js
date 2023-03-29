import {
  SET_NAVIGATED_FROM_DASHBOARD,
  SET_NAVIGATED_FROM_ANALYSE,
  SET_COMPARISON_ENABLED,
  COMPARISON_DATA_LOADING,
  COMPARISON_DATA_FETCHED,
  COMPARISON_DATA_ERROR,
  RESET_COMPARISON_DATA,
  SET_COMPARISON_SUPPORTED,
  SET_COMPARE_DURATION,
  UPDATE_CHART_TYPES,
  SET_SAVED_QUERY_SETTINGS,
  UPDATE_PIVOT_CONFIG,
  UPDATE_FUNNEL_TABLE_CONFIG,
  UPDATE_CORE_QUERY_REDUCER
} from './constants';

export default function (state, action) {
  const { payload } = action;
  switch (action.type) {
    case UPDATE_CORE_QUERY_REDUCER: {
      return {
        ...state,
        ...payload
      };
    }
    case SET_NAVIGATED_FROM_DASHBOARD:
      return {
        ...state,
        navigatedFromDashboard: payload
      };
    case SET_NAVIGATED_FROM_ANALYSE:
      return {
        ...state,
        navigatedFromAnalyse: payload
      };
    case SET_COMPARISON_SUPPORTED:
      return {
        ...state,
        comparison_supported: payload
      };
    case SET_COMPARISON_ENABLED:
      return {
        ...state,
        comparison_enabled: payload
      };
    case COMPARISON_DATA_LOADING:
      return {
        ...state,
        comparison_data: {
          loading: true,
          error: false,
          data: null
        }
      };
    case COMPARISON_DATA_FETCHED:
      return {
        ...state,
        comparison_data: {
          loading: false,
          error: false,
          data: payload
        }
      };
    case COMPARISON_DATA_ERROR:
      return {
        ...state,
        comparison_data: {
          loading: false,
          error: true,
          data: null
        }
      };
    case RESET_COMPARISON_DATA:
      return {
        ...state,
        comparison_duration: null,
        comparison_enabled: false,
        comparison_data: {
          loading: false,
          error: false,
          data: null
        }
      };
    case SET_COMPARE_DURATION:
      return {
        ...state,
        comparison_duration: payload
      };
    case UPDATE_CHART_TYPES:
      return {
        ...state,
        chartTypes: payload
      };
    case SET_SAVED_QUERY_SETTINGS: {
      return {
        ...state,
        savedQuerySettings: payload
      };
    }
    case UPDATE_PIVOT_CONFIG: {
      return {
        ...state,
        pivotConfig: {
          ...state.pivotConfig,
          ...payload
        }
      };
    }
    case UPDATE_FUNNEL_TABLE_CONFIG: {
      return {
        ...state,
        funnelTableConfig: payload
      };
    }
    default:
      return state;
  }
}
