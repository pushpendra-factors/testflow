import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { get, getHostUrl, post, del, put } from '../../utils/request';
import {
  DASHBOARDS_LOADED,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADED
} from '../types';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import { changeActiveDashboardAction } from './actions';

const host = getHostUrl();

export const fetchDashboards = (projectId) => {
  return async function (dispatch) {
    try {
      dispatch({ type: DASHBOARDS_LOADING });
      const url = host + 'projects/' + projectId + '/dashboards';
      const res = await get(null, url);
      dispatch({ type: DASHBOARDS_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: DASHBOARDS_LOADING_FAILED });
    }
  };
};

export const saveQueryToDashboard = (
  projectId,
  selectedDashboardIds,
  reqBody
) => {
  const url =
    host +
    'projects/' +
    projectId +
    '/v1/dashboards/multi/' +
    selectedDashboardIds +
    '/units';
  return post(null, url, reqBody);
};

export const fetchActiveDashboardUnits = (projectId, activeDashboardId) => {
  return async function (dispatch) {
    try {
      dispatch({ type: DASHBOARD_UNITS_LOADING });
      const url =
        host +
        'projects/' +
        projectId +
        '/dashboards/' +
        activeDashboardId +
        '/units';
      const res = await get(null, url);
      dispatch({ type: DASHBOARD_UNITS_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: DASHBOARD_UNITS_LOADING_FAILED });
    }
  };
};

export const createDashboard = async (projectId, reqBody) => {
  const url = host + 'projects/' + projectId + '/dashboards';
  return post(null, url, reqBody);
};

export const assignUnitsToDashboard = async (
  projectId,
  dashboardId,
  reqBody
) => {
  const url =
    host +
    'projects/' +
    projectId +
    '/v1/dashboards/queries/' +
    dashboardId +
    '/units';
  return post(null, url, reqBody);
};

export const deleteDashboard = (projectId, dashboardId) => {
  const url = host + 'projects/' + projectId + '/v1/dashboards/' + dashboardId;
  return del(null, url);
};

export const updateDashboard = (projectId, dashboardId, body) => {
  const url = host + 'projects/' + projectId + '/dashboards/' + dashboardId;
  return put(null, url, body);
};

export const DeleteUnitFromDashboard = (projectId, dashboardId, unitId) => {
  const url =
    host +
    'projects/' +
    projectId +
    '/dashboards/' +
    dashboardId +
    '/units/' +
    unitId;
  return del(null, url);
};

export const changeActiveDashboard = (selectedDashboard) => {
  return function (dispatch) {
    setItemToLocalStorage(
      DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID,
      selectedDashboard.id
    );
    dispatch(changeActiveDashboardAction(selectedDashboard));
  };
};
