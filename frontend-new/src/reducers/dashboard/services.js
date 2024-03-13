import { notification } from 'antd';
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

import {
  DASHBOARD_FOLDERS_LIST_ERROR,
  DASHBOARD_FOLDERS_LIST_LOADING,
  DASHBOARD_FOLDERS_LIST_SUCCESS,
  DASHBOARD_NEW_FOLDER_CREATION_SUCCESSFUL,
  INITIATED_DASHBOARD_NEW_FOLDER_CREATION,
  INITIATED_DASHBOARD_MOVE_TO_EXISTING_FOLDER,
  DASHBOARD_MOVE_TO_EXISTING_FOLDER_SUCCESSFUL,
  DASHBOARD_MOVE_TO_EXISTING_FOLDER_FAILED,
  DASHBOARD_NEW_FOLDER_CREATION_FAILED,
  INITIATED_RENAME_DASHBOARD_FOLDER,
  RENAME_DASHBOARD_FOLDER_SUCCESSFUL,
  RENAME_DASHBOARD_FOLDER_FAILED,
  INITIATED_DELETE_DASHBOARD_FOLDER,
  DELETE_DASHBOARD_FOLDER_SUCCESSFUL,
  DELETE_DASHBOARD_FOLDER_FAILED
} from './types';

const host = getHostUrl();

export const fetchDashboards = (projectId) =>
  async function (dispatch) {
    try {
      dispatch({ type: DASHBOARDS_LOADING });
      const url = `${host}projects/${projectId}/dashboards`;
      const res = await get(null, url);
      dispatch({ type: DASHBOARDS_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: DASHBOARDS_LOADING_FAILED });
    }
  };

export const saveQueryToDashboard = (
  projectId,
  selectedDashboardIds,
  reqBody
) => {
  const url = `${host}projects/${projectId}/v1/dashboards/multi/${selectedDashboardIds}/units`;
  return post(null, url, reqBody);
};

export const fetchActiveDashboardUnits = (projectId, activeDashboardId) =>
  async function (dispatch) {
    try {
      dispatch({ type: DASHBOARD_UNITS_LOADING });
      const url = `${host}projects/${projectId}/dashboards/${activeDashboardId}/units`;
      const res = await get(null, url);
      dispatch({ type: DASHBOARD_UNITS_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: DASHBOARD_UNITS_LOADING_FAILED });
    }
  };

export const createDashboard = async (projectId, reqBody) => {
  const url = `${host}projects/${projectId}/dashboards`;
  return post(null, url, reqBody);
};

export const assignUnitsToDashboard = async (
  projectId,
  dashboardId,
  reqBody
) => {
  const url = `${host}projects/${projectId}/v1/dashboards/queries/${dashboardId}/units`;
  return post(null, url, reqBody);
};

export const deleteDashboard = (projectId, dashboardId) => {
  const url = `${host}projects/${projectId}/v1/dashboards/${dashboardId}`;
  return del(null, url);
};

export const updateDashboard = (projectId, dashboardId, body) => {
  const url = `${host}projects/${projectId}/dashboards/${dashboardId}`;
  return put(null, url, body);
};

export const DeleteUnitFromDashboard = (projectId, dashboardId, unitId) => {
  const url = `${host}projects/${projectId}/dashboards/${dashboardId}/units/${unitId}`;
  return del(null, url);
};

export const changeActiveDashboard = (selectedDashboard) =>
  function (dispatch) {
    setItemToLocalStorage(
      DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID,
      selectedDashboard.id
    );
    dispatch(changeActiveDashboardAction(selectedDashboard));
  };

export const fetchDashboardFolders = (projectId) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/dashboard_folder`;
      dispatch({ type: DASHBOARD_FOLDERS_LIST_LOADING });
      const res = await get(null, url);
      dispatch({
        type: DASHBOARD_FOLDERS_LIST_SUCCESS,
        payload: res.data
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: DASHBOARD_FOLDERS_LIST_ERROR
      });
    }
  };

export const addDashboardToNewFolder = (projectId, dashboardId, folderName) =>
  async function (dispatch) {
    try {
      dispatch({ type: INITIATED_DASHBOARD_NEW_FOLDER_CREATION });
      const url = `${host}projects/${projectId}/dashboard_folder`;
      const res = await post(null, url, {
        name: folderName,
        dashboard_id: Number(dashboardId)
      });
      notification.success({
        message: 'Success',
        description: 'Folder creation successful',
        duration: 2
      });
      dispatch({
        type: DASHBOARD_NEW_FOLDER_CREATION_SUCCESSFUL,
        payload: { newFolder: res.data, dashboardId }
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: DASHBOARD_NEW_FOLDER_CREATION_FAILED
      });
      notification.error({
        message: 'Error',
        description: 'Folder creation failed',
        duration: 2
      });
    }
  };

export const addDashboardToExistingFolder = (
  projectId,
  folderId,
  dashboardId
) =>
  async function (dispatch) {
    try {
      dispatch({ type: INITIATED_DASHBOARD_MOVE_TO_EXISTING_FOLDER });
      const url = `${host}projects/${projectId}/dashboards/${dashboardId}`;
      await put(null, url, {
        folder_id: folderId
      });
      notification.success({
        message: 'Success',
        description: 'Dashboard moved successfully',
        duration: 2
      });
      dispatch({
        type: DASHBOARD_MOVE_TO_EXISTING_FOLDER_SUCCESSFUL,
        payload: { folderId, dashboardId }
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: DASHBOARD_MOVE_TO_EXISTING_FOLDER_FAILED
      });
      notification.error({
        message: 'Error',
        description: 'Dashboard move failed',
        duration: 2
      });
    }
  };

export const renameDashboardFolder = (projectId, folderId, newName) =>
  async function (dispatch) {
    try {
      dispatch({ type: INITIATED_RENAME_DASHBOARD_FOLDER });
      const url = `${host}projects/${projectId}/dashboard_folder/${folderId}`;
      await put(null, url, {
        name: newName
      });
      notification.success({
        message: 'Success',
        description: 'Folder rename successful',
        duration: 2
      });
      dispatch({
        type: RENAME_DASHBOARD_FOLDER_SUCCESSFUL,
        payload: {
          folderId,
          newName
        }
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: RENAME_DASHBOARD_FOLDER_FAILED
      });
      notification.error({
        message: 'Error',
        description: 'Folder rename failed',
        duration: 2
      });
    }
  };

export const deleteDashboardFolder = (projectId, folderId) =>
  async function (dispatch) {
    try {
      dispatch({ type: INITIATED_DELETE_DASHBOARD_FOLDER });
      const url = `${host}projects/${projectId}/dashboards/folder/${folderId}`;
      await put(null, url);
      const deleteUrl = `${host}projects/${projectId}/dashboard_folder/${folderId}`;
      await del(null, deleteUrl);
      notification.success({
        message: 'Success',
        description: 'Folder deleted successfully',
        duration: 2
      });
      dispatch({
        type: DELETE_DASHBOARD_FOLDER_SUCCESSFUL,
        payload: {
          folderId
        }
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: DELETE_DASHBOARD_FOLDER_FAILED
      });
      notification.error({
        message: 'Error',
        description: 'Folder deletion failed',
        duration: 2
      });
    }
  };
