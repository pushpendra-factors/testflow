import { getHostUrl, post, del } from '../../utils/request';
import {
  DASHBOARDS_LOADED,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADED
} from '../types';

const host = getHostUrl();

export const fetchDashboards = (projectId) => {
  return async function (dispatch) {
    try {
      dispatch({ type: DASHBOARDS_LOADING });
      const url = 'projects/' + projectId + '/dashboards';
      const res = await post(null, host, { url, method: 'GET' });
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
    'projects/' +
    projectId +
    '/v1/dashboards/multi/' +
    selectedDashboardIds +
    '/units';
  return post(null, host, { url, requestBody: reqBody, method: 'POST' });
};

export const fetchActiveDashboardUnits = (projectId, activeDashboardId) => {
  return async function (dispatch) {
    try {
      dispatch({ type: DASHBOARD_UNITS_LOADING });
      const url =
        'projects/' + projectId + '/dashboards/' + activeDashboardId + '/units';
      const res = await post(null, host, { url, method: 'GET' });
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
  const url = 'projects/' + projectId + '/dashboards/' + dashboardId;
  return post(null, host, { url, method: 'PUT', requestBody: body });
};

export const DeleteUnitFromDashboard = (projectId, dashboardId, unitId) => {
  const url =
    'projects/' + projectId + '/dashboards/' + dashboardId + '/units/' + unitId;
  return post(null, host, { url, method: 'DELETE' });
};
