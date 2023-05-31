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
    case 'FETCH_SMART_EVENTS_FULFILLED': {
      return { ...state, smart_events: action.payload };
    }
    case 'FETCH_SMART_EVENTS_REJECTED': {
      return { ...state, error: action.payload };
    }
    case 'FETCH_OBJECTPROPERTIESBYSOURCE_FULFILLED': {
      return { ...state, objPropertiesSource: action.payload };
    }
    case 'FETCH_OBJECTPROPERTIESBYSOURCE_REJECTED': {
      return { ...state, error: action.payload };
    }
    case 'SAVE_SMART_EVENTS_FULFILLED': {
      return { ...state };
    }
    case 'SAVE_SMART_EVENTS_REJECTED': {
      return { ...state, error: action.payload };
    }
    case 'FETCH_SPECIFICPROPERTIESVALUE_FULFILLED': {
      return { ...state, specificPropertiesData: action.payload };
    }
    case 'FETCH_SPECIFICPROPERTIESVALUE_REJECTED': {
      return { ...state, error: action.payload };
    }
    case 'ENABLE_ADWORDS_FULFILLED': {
      let enabledAgentUUID = action.payload.int_adwords_enabled_agent_uuid;
      if (!enabledAgentUUID || enabledAgentUUID === '') return state;

      let _state = { ...state };
      _state.currentProjectSettings = {
        ...state.currentProjectSettings,
        int_adwords_enabled_agent_uuid: enabledAgentUUID
      };
      return _state;
    }
    case 'FETCH_ADWORDS_CUSTOMER_ACCOUNTS_FULFILLED': {
      let _state = { ...state };
      _state.adwordsCustomerAccounts = [...action.payload.customer_accounts];
      return _state;
    }
    case SET_ACTIVE_PROJECT: {
      return {
        ...initialState
      };
    }
    default:
      return state;
  }
}

export function fetchSmartEvents(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectID + '/v1/smart_event')
        .then((response) => {
          dispatch({
            type: 'FETCH_SMART_EVENTS_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_SMART_EVENTS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function removeSmartEvents(projectID, filterID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      del(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/smart_event?filter_id=' +
          filterID +
          '&type=crm'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_SMART_EVENTS_REMOVE_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_SMART_EVENTS_REMOVE_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}

export function saveSmartEvents(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectID + '/v1/smart_event?type=crm',
        data
      )
        .then((response) => {
          dispatch({
            type: 'SAVE_SMART_EVENTS_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'SAVE_SMART_EVENTS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchObjectPropertiesbySource(projectID, source, dataObj) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/crm/' +
          source +
          '/' +
          dataObj +
          '/properties'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_OBJECTPROPERTIESBYSOURCE_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_OBJECTPROPERTIESBYSOURCE_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}

export function fetchSpecificPropertiesValue(
  projectID,
  source,
  dataObj,
  propName
) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/v1/crm/' +
          source +
          '/' +
          dataObj +
          '/properties/' +
          propName +
          '/values'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_SPECIFICPROPERTIESVALUE_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_SPECIFICPROPERTIESVALUE_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}
