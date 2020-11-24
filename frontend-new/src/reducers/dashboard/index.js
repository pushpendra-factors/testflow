import {
  DASHBOARDS_LOADED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADED,
  ACTIVE_DASHBOARD_CHANGE,
  DASHBOARD_UNIT_DATA_LOADED,
  DASHBOARD_CREATED,
  DASHBOARD_DELETED,
  UNITS_ORDER_CHANGED,
  DASHBOARD_UNMOUNTED
} from '../types';
import { getRearrangedData } from './utils';

const defaultState = {
  dashboardsLoaded: 0,
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
      return { ...state, activeDashboardUnits: { ...defaultState.activeDashboardUnits, data: getRearrangedData(action.payload, state.activeDashboard) } };
    case ACTIVE_DASHBOARD_CHANGE:
      return { ...state, activeDashboard: action.payload, activeDashboardUnits: { ...defaultState.activeDashboardUnits } };
    case DASHBOARD_UNIT_DATA_LOADED:
      return { ...state, dashboardsLoaded: state.dashboardsLoaded + 1 };
    case DASHBOARD_CREATED:
      return { ...state, dashboards: { ...state.dashboards, data: [...state.dashboards.data, action.payload] } };
    case DASHBOARD_DELETED: {
      const newDashboardList = state.dashboards.data.filter(d => d.id !== action.payload.id);
      const newActiveDashboard = newDashboardList[0];
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
        dashboards: { ...defaultState.dashboards, data: newDashboardList },
        activeDashboard: newActiveDashboard
      };
    }
    case UNITS_ORDER_CHANGED: {
      return {
        ...state,
        activeDashboardUnits: {
          ...state.activeDashboardUnits,
          data: [...action.payload]
        },
        dashboardsLoaded: state.dashboardsLoaded + 1
      };
    }
    case DASHBOARD_UNMOUNTED:
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits }
      };
    default:
      return state;
  }
}
