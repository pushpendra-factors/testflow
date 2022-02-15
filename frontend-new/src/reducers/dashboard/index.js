import { isNull } from "lodash";
import {
  DASHBOARDS_LOADED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADED,
  ACTIVE_DASHBOARD_CHANGE,
  DASHBOARD_CREATED,
  DASHBOARD_DELETED,
  UNITS_ORDER_CHANGED,
  DASHBOARD_UNMOUNTED,
  WIDGET_DELETED,
  DASHBOARD_UPDATED,
  SET_ACTIVE_PROJECT,
  DASHBOARD_LAST_REFRESHED,
} from "../types";
import { getRearrangedData } from "./utils";

const defaultState = {
  dashboards: {
    loading: false,
    error: false,
    data: [],
  },
  activeDashboard: {},
  activeDashboardUnits: {
    loading: false,
    error: false,
    data: [],
  },
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case DASHBOARDS_LOADING:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, loading: true },
      };
    case DASHBOARDS_LOADING_FAILED:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, error: true },
      };
    case DASHBOARDS_LOADED:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, data: action.payload },
        activeDashboard: isNull(localStorage.getItem('active-dashboard-id'))? action.payload[0]:JSON.parse(localStorage.getItem('active-dashboard-id')),
      };
    case DASHBOARD_UNITS_LOADING:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          loading: true,
        },
      };
    case DASHBOARD_UNITS_LOADING_FAILED:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          error: true,
        },
      };
    case DASHBOARD_UNITS_LOADED:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          data: getRearrangedData(action.payload, state.activeDashboard),
        },
      };
    case ACTIVE_DASHBOARD_CHANGE:
      return {
        ...state,
        activeDashboard: action.payload,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
      };
    case DASHBOARD_LAST_REFRESHED:
      return {
        ...state,
        activeDashboard: {
          ...state.activeDashboard,
          refreshed_at: action.payload,
        },
      };
    case DASHBOARD_CREATED:
      return {
        ...state,
        dashboards: {
          ...state.dashboards,
          data: [...state.dashboards.data, action.payload],
        },
      };
    case DASHBOARD_DELETED: {
      const newDashboardList = state.dashboards.data.filter(
        (d) => d.id !== action.payload.id
      );
      const newActiveDashboard = newDashboardList[0];
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
        dashboards: { ...defaultState.dashboards, data: newDashboardList },
        activeDashboard: newActiveDashboard,
      };
    }
    case WIDGET_DELETED: {
      return {
        ...state,
        activeDashboardUnits: {
          ...state.activeDashboardUnits,
          data: state.activeDashboardUnits.data.filter(
            (elem) => elem.id !== action.payload
          ),
        },
      };
    }
    case UNITS_ORDER_CHANGED: {
      const activeDashboardIdx = state.dashboards.data.findIndex(
        (elem) => elem.id === state.activeDashboard.id
      );
      return {
        ...state,
        activeDashboardUnits: {
          ...state.activeDashboardUnits,
          data: [...action.payload],
        },
        activeDashboard: {
          ...state.activeDashboard,
          units_position: action.units_position,
        },
        dashboards: {
          ...state.dashboards,
          data: [
            ...state.dashboards.data.slice(0, activeDashboardIdx),
            { ...state.activeDashboard, units_position: action.units_position },
            ...state.dashboards.data.slice(activeDashboardIdx + 1),
          ],
        },
      };
    }
    case DASHBOARD_UNMOUNTED:
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
      };
    case DASHBOARD_UPDATED:
      const dashboardIndex = state.dashboards.data.findIndex(
        (dashboard) => dashboard.id === action.payload.id
      );
      const editedDashboard = {
        ...state.dashboards.data[dashboardIndex],
        ...action.payload,
      };
      return {
        ...state,
        activeDashboard: editedDashboard,
        dashboards: {
          ...state.dashboards,
          data: [
            ...state.dashboards.data.slice(0, dashboardIndex),
            editedDashboard,
            ...state.dashboards.data.slice(dashboardIndex + 1),
          ],
        },
      };
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState,
      };
    default:
      return state;
  }
}
