import { resolve } from 'path';
import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { get, getHostUrl, post, put } from '../../utils/request';

let host = getHostUrl();
host = host[host.length - 1] === '/' ? host : `${host}/`;

const initialState = {
  contacts: { isLoading: false, data: [] },
  contactDetails: { isLoading: false, data: {} },
  accounts: { isLoading: false, data: [] },
  accountDetails: { isLoading: false, data: {} },
  segmentCreateStatus: '',
  segmentUpdateStatus: '',
  segments: {}
};

export default function (state = initialState, action) {
  switch (action.type) {
    case 'FETCH_PROFILE_USERS_LOADING':
      return { ...state, contacts: { isLoading: true, data: [] } };
    case 'FETCH_PROFILE_USERS_FULFILLED':
      return { ...state, contacts: { isLoading: false, data: action.payload } };
    case 'FETCH_PROFILE_USERS_FAILED':
      return { ...state, contacts: { isLoading: false, data: [] } };
    case 'FETCH_PROFILE_USER_DETAILS_LOADING':
      return { ...state, contactDetails: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_USER_DETAILS_FULFILLED':
      return {
        ...state,
        contactDetails: { isLoading: false, data: action.payload }
      };
    case 'FETCH_PROFILE_USER_DETAILS_UPDATED':
      return {
        ...state,
        contactDetails: { isLoading: false, data: action.payload }
      };
    case 'FETCH_PROFILE_USER_DETAILS_FAILED':
      return { ...state, contactDetails: { isLoading: false, data: {} } };
    case 'FETCH_PROFILE_ACCOUNTS_LOADING':
      return { ...state, accounts: { isLoading: true, data: [] } };
    case 'FETCH_PROFILE_ACCOUNTS_FULFILLED':
      return { ...state, accounts: { isLoading: false, data: action.payload } };
    case 'FETCH_PROFILE_ACCOUNTS_FAILED':
      return { ...state, accounts: { isLoading: false, data: [] } };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING':
      return { ...state, accountDetails: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED':
      return {
        ...state,
        accountDetails: { isLoading: false, data: action.payload }
      };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED':
      return { ...state, accountDetails: { isLoading: false, data: {} } };
    case 'SEGMENT_CREATION_FULFILLED':
      return { ...state, segmentCreateStatus: action.payload };
    case 'SEGMENT_CREATION_REJECTED':
      return { ...state, segmentCreateStatus: action.payload };
    case 'FETCH_SEGMENTS_FULFILLED':
      return { ...state, segments: action.payload };
    case 'FETCH_SEGMENTS_REJECTED':
      return { ...state, segments: {} };
    case 'UPDATE_SEGMENT_FULFILLED':
      return { ...state, segmentUpdateStatus: action.payload };
    case 'UPDATE_SEGMENT_REJECTED':
      return { ...state, segmentUpdateStatus: action.payload };
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}

const getURLWithQueryParams = (projectId, profileType, agentId) => {
  const queryParams = [];

  if (window.SCORE || agentId === 'solutions@factors.ai') {
    queryParams.push('score=true');
  }

  if (window.SCORE_DEBUG) {
    queryParams.push('debug=true');
  }

  const queryString = queryParams.join('&');

  const url = `${host}projects/${projectId}/v1/profiles/${profileType}${
    queryString ? `?${queryString}` : ''
  }`;

  return url;
};

export const fetchProfileUsers = (projectId, reqBody, agentId) => {
  let url = getURLWithQueryParams(projectId, 'users', agentId);
  return post(null, url, reqBody);
};

export const fetchProfileUserDetails = (projectId, id, isAnonymous) => {
  const url = `${host}projects/${projectId}/v1/profiles/users/${id}?is_anonymous=${isAnonymous}`;
  return get(null, url);
};

export const fetchProfileAccounts = (projectId, reqBody, agentId) => {
  const url = getURLWithQueryParams(projectId, 'accounts', agentId);
  return post(null, url, reqBody);
};

export const fetchProfileAccountDetails = (projectId, id, group) => {
  const url = `${host}projects/${projectId}/v1/profiles/accounts/${group}/${id}`;
  return get(null, url);
};

export const createSegment = (projectId, payload) => {
  const url = `${host}projects/${projectId}/segments`;
  return post(null, url, payload);
};

export const fetchSegments = (projectId) => {
  const url = `${host}projects/${projectId}/segments`;
  return get(null, url);
};

export const fetchSegmentById = (projectId, id) => {
  const url = `${host}projects/${projectId}/segments/${id}`;
  return get(null, url);
};

export const updateSegment = (projectId, id, payload) => {
  const url = `${host}projects/${projectId}/segments/${id}`;
  return put(null, url, payload);
};

export const updateAccountScores = (projectID, payload) => {
  const url = `${host}projects/${projectID}/v1/accscore/weights`;
  return put(null, url, payload);
};
