import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { del, get, getHostUrl, post, put } from '../../utils/request';
import {
  LOADING_SEGMENT_FOLDER,
  SET_ACCOUNTS_SEGMENT_FOLDERS_FAILED,
  SET_ACCOUNT_SEGMENT_FOLDERS,
  SET_PEOPLES_SEGMENT_FOLDERS_FAILED,
  SET_PEOPLE_SEGMENT_FOLDERS
} from './types';

let host = getHostUrl();
host = host[host.length - 1] === '/' ? host : `${host}/`;

const initialState = {
  contacts: { isLoading: false, data: [] },
  contactDetails: { isLoading: false, data: {} },
  accounts: {},
  accountDetails: { isLoading: false, data: {} },
  accountOverview: { isLoading: false, data: {} },
  segmentCreateStatus: '',
  segmentUpdateStatus: '',
  segmentFolders: {
    isLoading: true,
    isSuccess: false,
    accounts: [],
    peoples: []
  },
  activePageView: '',
  accountPreview: {},
  userConfigProperties: {},
  eventConfigProperties: {},
  eventPropertiesType: {},
  accountSegments: [],
  userSegments: []
};

const updateAccountState = (state, segmentID, updates) => ({
  ...state,
  accounts: {
    ...state.accounts,
    [segmentID]: {
      ...state.accounts[segmentID],
      ...updates
    }
  }
});

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
      return updateAccountState(state, action.segmentID, { isLoading: true });
    case 'FETCH_PROFILE_ACCOUNTS_FULFILLED':
      return updateAccountState(state, action.segmentID, {
        isLoading: false,
        profiles: action.payload,
        isPreview: action.isPreview
      });
    case 'FETCH_PROFILE_ACCOUNTS_FAILED':
      return updateAccountState(state, action.segmentID, {
        isLoading: false,
        profiles: [],
        isPreview: false
      });

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
      return {
        ...state,
        accountSegments: action.accountSegments,
        userSegments: action.userSegments
      };
    case 'FETCH_SEGMENTS_REJECTED':
      return { ...state, accountSegments: [], userSegments: [] };
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
    case SET_ACCOUNT_SEGMENT_FOLDERS:
      return {
        ...state,
        segmentFolders: {
          ...state.segmentFolders,
          isLoading: false,
          isSuccess: true,
          accounts: action.payload
        }
      };
    case SET_PEOPLE_SEGMENT_FOLDERS:
      return {
        ...state,
        segmentFolders: {
          ...state.segmentFolders,
          isLoading: false,
          isSuccess: true,
          peoples: action.payload
        }
      };
    case SET_ACCOUNTS_SEGMENT_FOLDERS_FAILED:
      return {
        ...state,
        segmentFolders: {
          ...state.segmentFolders,
          isLoading: false,
          isSuccess: false,
          accounts: []
        }
      };
    case SET_PEOPLES_SEGMENT_FOLDERS_FAILED:
      return {
        ...state,
        segmentFolders: {
          ...state.segmentFolders,
          isLoading: false,
          isSuccess: false,
          peoples: []
        }
      };
    case LOADING_SEGMENT_FOLDER:
      return {
        ...state,
        segmentFolders: {
          ...state.segmentFolders,
          isLoading: true
        }
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

export const updateTableProperties = (projectID, profileType, payload) => {
  const url = `${host}projects/${projectID}/v1/profiles/${profileType}/table_properties`;
  return put(null, url, payload);
};

export const updateTablePropertiesForSegment = (
  projectID,
  segmentID,
  payload
) => {
  const url = `${host}projects/${projectID}/v1/profiles/segments/${segmentID}/table_properties`;
  return put(null, url, payload);
};

export const fetchSegmentFolders = (projectID, folder_type) => {
  const url = `${host}projects/${projectID}/segment_folders?type=${folder_type}`;
  return get(null, url);
};
export const renameSegmentFolders = (
  projectID,
  folderID,
  payload,
  folder_type
) => {
  const url = `${host}projects/${projectID}/segment_folders/${folderID}?type=${folder_type}`;
  return put(null, url, payload);
};
export const deleteSegmentFolders = (projectID, folderID, folder_type) => {
  const url = `${host}projects/${projectID}/segment_folders/${folderID}?type=${folder_type}`;
  return del(null, url);
};
// Move segment(:id) to Folder(with id) which needs to be passed via body
// payload = {folder_id: string | number}
export const updateSegmentToFolder = (
  projectID,
  segmentID,
  payload,
  folder_type
) => {
  const url = `${host}projects/${projectID}/segment_folders_item/${segmentID}?type=${folder_type}`;
  return put(null, url, payload);
};

// Move segment(:id) to Folder(with id) which needs to be passed via body
// payload = {folder_id: string | number}
export const moveSegmentToNewFolder = (
  projectID,
  segmentID,
  payload,
  folder_type
) => {
  const url = `${host}projects/${projectID}/segment_folders_item/${segmentID}?type=${folder_type}`;
  return post(null, url, payload);
};
