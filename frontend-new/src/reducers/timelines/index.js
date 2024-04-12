import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { del, get, getHostUrl, post, put } from '../../utils/request';
import { SEGMENT_DELETED } from './types';

let host = getHostUrl();
host = host[host.length - 1] === '/' ? host : `${host}/`;

const initialState = {
  contacts: { isLoading: false, data: [] },
  contactDetails: { isLoading: false, data: {} },
  accounts: { isLoading: true, data: {} },
  accountDetails: { isLoading: false, data: {} },
  accountOverview: { isLoading: false, data: {} },
  segmentCreateStatus: '',
  segmentUpdateStatus: '',
  segments: {},
  activePageView: '',
  accountPreview: {},
  userConfigProperties: {},
  eventConfigProperties: {},
  eventPropertiesType: {}
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
      return { ...state, accounts: { ...state.accounts, isLoading: true } };

    case 'FETCH_PROFILE_ACCOUNTS_FULFILLED':
      const updatedData = { ...state.accounts.data };
      updatedData[action.segmentID || 'default'] = action.payload;
      return { ...state, accounts: { isLoading: false, data: updatedData } };
    case 'FETCH_PROFILE_ACCOUNTS_FAILED':
      return { ...state, accounts: { ...state.accounts, isLoading: false } };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING':
      return { ...state, accountDetails: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED':
      return {
        ...state,
        accountDetails: { isLoading: false, data: action.payload }
      };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED':
      return { ...state, accountDetails: { isLoading: false, data: {} } };
    case 'FETCH_PROFILE_ACCOUNT_OVERVIEW_LOADING':
      return { ...state, accountOverview: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FULFILLED':
      return {
        ...state,
        accountOverview: { isLoading: false, data: action.payload }
      };
    case 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FAILED':
      return { ...state, accountOverview: { isLoading: false, data: {} } };
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
    case 'SET_PAGEVIEW':
      return { ...state, activePageView: action.payload };
    case 'FETCH_TOP100_EVENTS_LOADING':
      return {
        ...state,
        accountPreview: {
          ...state.accountPreview,
          [action.domainName]: { loading: true }
        }
      };
    case 'FETCH_TOP100_EVENTS_FULFILLED':
      return {
        ...state,
        accountPreview: {
          ...state.accountPreview,
          [action.domainName]: {
            loading: false,
            events: action.payload
          }
        }
      };
    case 'FETCH_TOP100_EVENTS_FAILED':
      return {
        ...state,
        accountPreview: {
          ...state.accountPreview,
          [action.domainName]: {
            loading: false,
            events: []
          }
        }
      };
    case 'FETCH_USER_CONFIG_PROPERTIES_FULFILLED':
      return {
        ...state,
        userConfigProperties: {
          ...state.userConfigProperties,
          [action.userID]: action.payload
        }
      };
    case 'FETCH_EVENT_CONFIG_PROPERTIES_FULFILLED':
      return {
        ...state,
        eventConfigProperties: {
          ...state.eventConfigProperties,
          [action.eventID]: action.payload
        }
      };
    case 'FETCH_USER_CONFIG_PROPERTIES_MAP_FULFILLED':
      return {
        ...state,
        userConfigProperties: {
          ...state.userConfigProperties,
          ...action.payload
        }
      };
    case 'FETCH_EVENT_CONFIG_PROPERTIES_MAP_FULFILLED':
      return {
        ...state,
        eventConfigProperties: {
          ...state.eventConfigProperties,
          ...action.payload
        }
      };
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    case SEGMENT_DELETED:
      return {
        ...state,
        segments: getUpdatedSegmentsAfterDeleting({
          segments: state.segments,
          segmentId: action.payload
        })
      };
    default:
      return state;
  }
}

const getURLWithQueryParams = (projectId, profileType, download) => {
  const queryParams = [];

  queryParams.push('user_marker=true');
  if (download) {
    queryParams.push('download=true');
  }

  const queryString = queryParams.join('&');

  const url = `${host}projects/${projectId}/v1/profiles/${profileType}${
    queryString ? `?${queryString}` : ''
  }`;

  return url;
};

export const fetchProfileUsers = (projectId, reqBody) => {
  const url = `${host}projects/${projectId}/v1/profiles/users`;
  return post(null, url, reqBody);
};

export const fetchProfileUserDetails = (projectId, id, isAnonymous) => {
  const url = `${host}projects/${projectId}/v1/profiles/users/${id}?is_anonymous=${isAnonymous}`;
  return get(null, url);
};

export const fetchProfileAccounts = (projectId, reqBody, download = false) => {
  const url = getURLWithQueryParams(projectId, 'accounts', download);
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

export const deleteSegmentByID = ({ projectId, segmentId }) => {
  const url = `${host}projects/${projectId}/segments/${segmentId}`;
  return del(null, url);
};

function getUpdatedSegmentsAfterDeleting({ segments, segmentId }) {
  return Object.fromEntries(
    Object.entries(segments).map(([key, list]) => [
      key,
      list.filter((segment) => segment.id !== segmentId)
    ])
  );
}

export const fetchAccountOverview = (projectID, groupName, accID) => {
  const url = `${host}projects/${projectID}/v1/profiles/accounts/overview/${groupName}/${accID}`;
  return get(null, url);
};

export const updateEngagementCategoryRanges = (projectID, payload) => {
  const url = `${host}projects/${projectID}/v1/accscore/engagementbuckets`;
  return put(null, url, payload);
};

export const getEngagementCategoryRanges = (projectID) => {
  const url = `${host}projects/${projectID}/v1/accscore/engagementbuckets`;
  return get(null, url);
};

export const updateEventPropertiesConfig = (projectID, eventName, payload) => {
  const url = `${host}projects/${projectID}/v1/profiles/events_config/${eventName}`;
  return put(null, url, payload);
};

export const fetchTop100Events = (projectID, domainName) => {
  const url = `${host}projects/${projectID}/v1/profiles/accounts/top_events/${btoa(
    domainName
  )}`;
  return get(null, url);
};

export const fetchConfiguredUserProperties = (
  projectID,
  userID,
  isAnonymous
) => {
  const url = `${host}projects/${projectID}/v1/profiles/user_properties/${userID}?is_anonymous=${isAnonymous}`;
  return get(null, url);
};

export const fetchConfiguredEventProperties = (
  projectID,
  eventID,
  eventName
) => {
  const url = `${host}projects/${projectID}/v1/profiles/event_properties/${eventID}/${eventName}`;
  return get(null, url);
};
