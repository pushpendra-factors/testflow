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
    case 'FETCH_TEMPLATE_CONFIG_FULFILLED': {
      return { ...state, config: action.payload };
    }
    case 'FETCH_TEMPLATE_INSIGHT_FULFILLED': {
      return { ...state, insight: action.payload };
    }
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}

export function fetchTemplateConfig(projectID, templateID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/templates/' +
          templateID +
          '/config'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_TEMPLATE_CONFIG_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_TEMPLATE_CONFIG_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchTemplateInsights(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectID + '/v1/templates/1/query',
        data
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_TEMPLATE_INSIGHT_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_TEMPLATE_INSIGHT_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
