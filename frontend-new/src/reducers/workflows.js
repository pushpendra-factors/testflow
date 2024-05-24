import { get, post, del, getHostUrl, put } from '../utils/request';
import { SET_ACTIVE_PROJECT } from './types';
var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  loading: false,
  error: false
};

export default function reducer(state = initialState, action) {
  switch (action.type) {
    case 'FETCH_SAVED_WORKFLOWS_FULFILLED': {
      return { ...state, savedWorkflows: action.payload };
    }
    case 'FETCH_WORKFLOW_TEMPLATES_FULFILLED': {
      return { ...state, templates: action.payload };
    }
    default:
      return state;
  }
}

export function fetchWorkflowTemplates(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/workflow/templates')
        .then((response) => {
          dispatch({
            type: 'FETCH_WORKFLOW_TEMPLATES_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_WORKFLOW_TEMPLATES_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function fetchSavedWorkflows(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/workflow/saved')
        .then((response) => {
          dispatch({
            type: 'FETCH_SAVED_WORKFLOWS_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_SAVED_WORKFLOWS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function saveWorkflow(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects/' + projectID + '/v1/workflow', data)
        .then((response) => {
          dispatch({
            type: 'FETCH_WORKFLOW_SAVE_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_WORKFLOW_SAVE_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function updateWorkflow(projectID, id, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(
        dispatch,
        host + 'projects/' + projectID + '/v1/workflow/edit/' + id,
        data
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_WORKFLOW_UPDATE_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_WORKFLOW_UPDATE_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function removeSavedWorkflow(projectID, id) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectID + `/v1/workflow/` + id)
        .then((response) => {
          dispatch({
            type: 'SAVED_WORKFLOW_REMOVED_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'SAVED_WORKFLOW_REMOVED_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
