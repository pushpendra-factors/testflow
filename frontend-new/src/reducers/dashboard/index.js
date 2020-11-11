import {
  DASHBOARDS_LOADED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADED,
  ACTIVE_DASHBOARD_CHANGE,
  DASHBOARD_UNIT_DATA_LOADED,
  DASHBOARD_CREATED
} from '../types';

const defaultState = {
  dashboards_loaded: 0,
  dashboards: {
    loading: false, error: false, data: []
  },
  activeDashboard: {},
  activeDashboardUnits: {
    loading: false, error: false, data: []
  }
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case DASHBOARDS_LOADING:
      return { ...defaultState, dashboards: { ...defaultState.dashboards, loading: true } };
    case DASHBOARDS_LOADING_FAILED:
      return { ...defaultState, dashboards: { ...defaultState.dashboards, error: true } };
    case DASHBOARDS_LOADED:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, data: action.payload },
        activeDashboard: action.payload[0]
      };
    case DASHBOARD_UNITS_LOADING:
      return { ...state, activeDashboardUnits: { ...defaultState.activeDashboardUnits, loading: true } };
    case DASHBOARD_UNITS_LOADING_FAILED:
      return { ...state, activeDashboardUnits: { ...defaultState.activeDashboardUnits, error: true } };
    case DASHBOARD_UNITS_LOADED:
      return { ...state, activeDashboardUnits: { ...defaultState.activeDashboardUnits, data: action.payload } };
    case ACTIVE_DASHBOARD_CHANGE:
      return { ...state, activeDashboard: action.payload, activeDashboardUnits: { ...defaultState.activeDashboardUnits } };
    case DASHBOARD_UNIT_DATA_LOADED:
      return { ...state, dashboards_loaded: state.dashboards_loaded + 1 }
    case DASHBOARD_CREATED:
      return { ...state, dashboards: { ...state.dashboards, data: [...state.dashboards.data, action.payload] } }
    default:
      return state;
  }
}
