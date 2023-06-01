import { get, post, put, getHostUrl } from '../utils/request';
import { SET_ACTIVE_PROJECT } from './types';
var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  loading: false,
  error: false,
  metadata: {},
  weekly_insights: {},
  active_insight: {}
};

export default function reducer(state = initialState, action) {
  switch (action.type) {
    case 'RESET_WEEKLY_INSIGHTS': {
      return { ...state, weekly_insights: initialState.weekly_insights };
    }
    case 'SET_ACTIVE_INSIGHT': {
      return { ...state, active_insight: action.payload };
    }
    case 'FETCH_WEEKLY_INSIGHTS_FULLFILLED': {
      return { ...state, weekly_insights: action.payload };
    }
    case 'FETCH_WEEKLY_INSIGHTS_REJECTED': {
      return { ...state };
    }
    case 'FETCH_WEEKLY_INSIGHTS_METADATA_FULLFILLED': {
      return { ...state, metadata: action.payload };
    }
    case 'FETCH_WEEKLY_INSIGHTS_METADATA_REJECTED': {
      return { ...state };
    }
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}

export function fetchWeeklyIngishts(
  projectID,
  dashboardID,
  baseTime,
  startTime,
  isDashboard = true,
  kpi_index = 1
) {
  const queryURL = isDashboard ? 'dashboard_unit_id' : 'query_id';
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host +
          'projects/' +
          projectID +
          '/insights?' +
          queryURL +
          '=' +
          dashboardID +
          '&base_start_time=' +
          baseTime +
          '&comp_start_time=' +
          startTime +
          '&insights_type=w&number_of_records=11&kpi_index=' +
          kpi_index
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_WEEKLY_INSIGHTS_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_WEEKLY_INSIGHTS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function fetchWeeklyIngishtsMetaData(projectID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectID + '/weekly_insights_metadata'
      )
        .then((response) => {
          dispatch({
            type: 'FETCH_WEEKLY_INSIGHTS_METADATA_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_WEEKLY_INSIGHTS_METADATA_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}
export function updateInsightFeedback(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects/' + projectID + `/feedback`, data)
        .then((response) => {
          dispatch({
            type: 'UPDATE_INSIGHT_FEEDBACK_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'UPDATE_INSIGHT_FEEDBACK__REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function buildExplainInsights(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectID + `/v1/explain`, data)
        .then((response) => {
          dispatch({
            type: 'UPDATE_INSIGHT_FEEDBACK_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'UPDATE_INSIGHT_FEEDBACK__REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function buildPathAnalysis(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectID + `/v1/pathanalysis`, data)
        .then((response) => {
          dispatch({
            type: 'UPDATE_INSIGHT_FEEDBACK_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'UPDATE_INSIGHT_FEEDBACK__REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function buildWeeklyInsights(projectID, data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectID + `/v1/weeklyinsights`, data)
        .then((response) => {
          dispatch({
            type: 'UPDATE_INSIGHT_FEEDBACK_FULLFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'UPDATE_INSIGHT_FEEDBACK__REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
