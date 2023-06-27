import { get, post, del, getHostUrl } from '../utils/request';
import { SET_ACTIVE_PROJECT } from './types';
var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  loading: false,
  error: false
};

export default function reducer(state = initialState, action) {
  switch (action.type) {
    case 'FETCH_KPI_CONFIG_FULFILLED': {
      return { ...state, config: action.payload };
    }
    case 'FETCH_KPI_CONFIG_WITHOUT_DERIVED_KPI_FULFILLED': {
      return { ...state, config_without_derived_kpi: action.payload };
    }
    case 'FETCH_CUSTOM_KPI_CONFIG_FULFILLED': {
      return { ...state, custom_kpi_config: action.payload };
    }
    case 'FETCH_KPI_PROPERTYMAPPING_FULFILLED': {
      return { ...state, kpi_property_mapping: action.payload };
    }
    case 'FETCH_SAVED_CUSTOM_KPI_FULFILLED': {
      return { ...state, saved_custom_kpi: action.payload };
    }
    case 'DEL_CUSTOM_KPI_FULFILLED': {
      return { ...state };
    }
    case 'FETCH_KPI_QUERY_FULFILLED': {
      return { ...state, query_result: action.payload };
    }
    case 'FETCH_KPI_PAGEURLS_FULFILLED': {
      return { ...state, page_urls: action.payload };
    }
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}

export function fetchKPIConfig(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/kpi/config?include_derived_kpis=true'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_CONFIG_FULFILLED',
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

export function fetchKPIConfigWithoutDerivedKPI(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/kpi/config?include_derived_kpis=false'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_CONFIG_WITHOUT_DERIVED_KPI_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_KPI_CONFIG_WITHOUT_DERIVED_KPI_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}

export function fetchCustomKPIConfig(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectID + '/v1/custom_metrics/config/v1'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_CUSTOM_KPI_CONFIG_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_CUSTOM_KPI_CONFIG_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function fetchSavedCustomKPI(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/custom_metrics')
        .then((response) => {
          dispatch({
            type: 'FETCH_SAVED_CUSTOM_KPI_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_SAVED_CUSTOM_KPI_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function addNewCustomKPI(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectID + `/v1/custom_metrics`,
        data
      )
        .then((response) => {
          dispatch({
            type: 'ADD_CUSTOM_KPI_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'ADD_CUSTOM_KPI_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function removeCustomKPI(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      del(
        dispatch,
        host + 'projects/' + projectID + `/v1/custom_metrics/` + data
      )
        .then((response) => {
          dispatch({
            type: 'DEL_CUSTOM_KPI_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'DEL_CUSTOM_KPI_REJECTED', payload: err });
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
            payload: response.data
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
            payload: response.data
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

export function getKPIPropertyMappings(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host +
          'projects/' +
          projectID +
          `/v1/kpi/property_mappings/commom_properties`,
        data
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_KPI_PROPERTYMAPPING_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_KPI_PROPERTYMAPPING_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}
