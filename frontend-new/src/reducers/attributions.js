import { get, post, getHostUrl } from '../utils/request';
import { SET_ACTIVE_PROJECT } from './types';
var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  loading: false,
  error: false
};

export default function reducer(state = initialState, action) {
  switch (action.type) {
    case 'FETCH_ATTR_CONFIG_FULFILLED': {
      return { ...state, attr_config: action.payload };
    }
    case 'FETCH_SAVED_ATTR_FULFILLED': {
      return { ...state, saved_attr: action.payload };
    }
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}

export function fetchAttrConfig(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectID + '/v1/custom_metrics/config'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_ATTR_CONFIG_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_ATTR_CONFIG_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function fetchSavedAttr(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/custom_metrics')
        .then((response) => {
          dispatch({
            type: 'FETCH_SAVED_ATTR_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_SAVED_ATTR_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function addNewAttr(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectID + `/v1/custom_metrics`,
        data
      )
        .then((response) => {
          dispatch({
            type: 'ADD_ATTR_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'ADD_ATTR_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
