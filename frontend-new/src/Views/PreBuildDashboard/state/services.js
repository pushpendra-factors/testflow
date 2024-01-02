import { EMPTY_ARRAY, EMPTY_OBJECT } from 'Utils/global';
import { get, getHostUrl, post, del, put } from 'Utils/request';
import { getRearrangedData } from 'Reducers/dashboard/utils';
import {
  ACTIVE_PRE_DASHBOARD_CHANGE,
  SET_ACTIVE_PROJECT
} from 'Reducers/types';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { DASHBOARD_KEYS } from 'Constants/localStorage.constants';

const host = getHostUrl();

export const DASHBOARD_CONFIG_LOADING = 'DASHBOARD_CONFIG_LOADING';
export const DASHBOARD_CONFIG_LOADED = 'DASHBOARD_CONFIG_LOADED';
export const DASHBOARD_CONFIG_LOADING_FAILED =
  'DASHBOARD_CONFIG_LOADING_FAILED';
export const SET_FILTER_PAYLOAD = 'SET_FILTER_PAYLOAD';
export const SET_REPORT_FILTER_PAYLOAD = 'SET_REPORT_FILTER_PAYLOAD';

export const changeActivePreDashboardAction = (newActiveDashboard) => ({
  type: ACTIVE_PRE_DASHBOARD_CHANGE,
  payload: newActiveDashboard
});

export const setFilterPayloadAction = (payload) => ({
  type: SET_FILTER_PAYLOAD,
  payload
});

export const setReportFilterPayloadAction = (payload) => ({
  type: SET_REPORT_FILTER_PAYLOAD,
  payload
});

export const changeActivePreDashboard = (selectedDashboard) =>
  function (dispatch) {
    setItemToLocalStorage(
      DASHBOARD_KEYS.ACTIVE_PRE_DASHBOARD_ID,
      selectedDashboard.id
    );
    dispatch(changeActivePreDashboardAction(selectedDashboard));
  };

export const fetchActiveDashboardConfig = (projectId, activeDashboardId) =>
  async function (dispatch) {
    try {
      dispatch({ type: DASHBOARD_CONFIG_LOADING });
      const url = `${host}projects/${projectId}/v1/predefined_dashboards/${activeDashboardId}/config`;
      const res = await get(null, url);
      dispatch({ type: DASHBOARD_CONFIG_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: DASHBOARD_CONFIG_LOADING_FAILED });
    }
  };

export const getQueryData = (projectId, query, internalID) => {
  let url;
  url = `${host}projects/${projectId}/v1/predefined_dashboards/${internalID}/query`;
  return post(null, url, query);
};
