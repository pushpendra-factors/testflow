/* eslint-disable */

import { get, post, del, getHostUrl } from '../utils/request';
var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const inititalState = {
  loading: false,
  error: false,
};

export default function reducer(state = inititalState, action) {
  switch (action.type) {
    case 'FETCH_KPI_CONFIG_FULFILLED': {
      return { ...state, config: action.payload };
    }
    case 'FETCH_KPI_QUERY_FULFILLED': {
      return { ...state, query_result: action.payload };
    }
    case 'FETCH_KPI_PAGEURLS_FULFILLED': {
      return { ...state, page_urls: action.payload };
    }
  }
  return state;
}

export function fetchKPIConfig(projectID, templateID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/kpi/config')
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_CONFIG_FULFILLED',
            payload: response.data,
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

export function fetchKPIFilterValues(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectID + `/v1/kpi/filter_values`,
        data
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_FILTERVALUES_FULFILLED',
            payload: response.data,
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_KPI_FILTERVALUES_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchPageUrls(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectID + `/v1/event_names/page_views`
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_PAGEURLS_FULFILLED',
            payload: response.data,
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_KPI_PAGEURLS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
