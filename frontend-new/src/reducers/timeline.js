import { get, getHostUrl, post } from '../utils/request';

var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  contacts: [],
  contactDetails: { isLoading: false, data: {} },
  accounts: { isLoading: false, data: [] },
  accountDetails: { isLoading: false, data: {} },
  error: false,
};

export default function (state = initialState, action) {
  switch (action.type) {
    case 'FETCH_PROFILE_USERS_FULFILLED':
      return { ...state, contacts: action.payload };
    case 'FETCH_PROFILE_USERS_FAILED':
      return { ...initialState, error: true };
    case 'FETCH_PROFILE_USER_DETAILS_LOADING':
      return { ...state, contactDetails: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_USER_DETAILS_FULFILLED':
      return {
        ...state,
        contactDetails: { isLoading: false, data: action.payload },
      };
    case 'FETCH_PROFILE_USER_DETAILS_FAILED':
      return { ...initialState, error: true };
    case 'FETCH_PROFILE_ACCOUNTS_LOADING':
      return { ...state, accounts: { isLoading: true, data: [] } };
    case 'FETCH_PROFILE_ACCOUNTS_FULFILLED':
      return { ...state, accounts: { isLoading: false, data: action.payload } };
    case 'FETCH_PROFILE_ACCOUNTS_FAILED':
      return { ...initialState, error: true };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING':
      return { ...state, accountDetails: { isLoading: true, data: {} } };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED':
      return {
        ...state,
        accountDetails: { isLoading: false, data: action.payload },
      };
    case 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED':
      return { ...initialState, error: true };
    default:
      return state;
  }
}

export const fetchProfileUsers = (projectId, reqBody) => {
  return async (dispatch) => {
    try {
      const url = host + 'projects/' + projectId + '/v1/profiles/users';
      const response = await post(null, url, reqBody);
      dispatch({
        type: 'FETCH_PROFILE_USERS_FULFILLED',
        payload: response.data,
      });
    } catch (err) {
      console.log(err);
      dispatch({ type: 'FETCH_PROFILE_USERS_FAILED' });
    }
  };
};

export const fetchProfileUserDetails = (projectId, id, isAnonymous) => {
  return async (dispatch) => {
    try {
      dispatch({ type: 'FETCH_PROFILE_USER_DETAILS_LOADING' });
      const url =
        host +
        'projects/' +
        projectId +
        '/v1/profiles/users/' +
        id +
        '?is_anonymous=' +
        isAnonymous;
      const response = await get(null, url);
      dispatch({
        type: 'FETCH_PROFILE_USER_DETAILS_FULFILLED',
        payload: response.data,
      });
    } catch (err) {
      console.log(err);
      dispatch({ type: 'FETCH_PROFILE_USER_DETAILS_FAILED' });
    }
  };
};

export const fetchProfileAccounts = (projectId, reqBody) => {
  return async (dispatch) => {
    try {
      dispatch({ type: 'FETCH_PROFILE_ACCOUNTS_LOADING' });
      const url = host + 'projects/' + projectId + '/v1/profiles/accounts';
      const response = await post(null, url, reqBody);
      dispatch({
        type: 'FETCH_PROFILE_ACCOUNTS_FULFILLED',
        payload: response.data,
      });
    } catch (err) {
      console.log(err);
      dispatch({ type: 'FETCH_PROFILE_ACCOUNTS_FAILED' });
    }
  };
};

export const fetchProfileAccountDetails = (projectId, id) => {
  return async (dispatch) => {
    try {
      dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
      const url =
        host + 'projects/' + projectId + '/v1/profiles/accounts/' + id;
      const response = await get(null, url);
      dispatch({
        type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
        payload: response.data,
      });
    } catch (err) {
      console.log(err);
      dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED' });
    }
  };
};
