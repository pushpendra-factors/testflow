import { get, getHostUrl, post } from '../utils/request';

var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const initialState = {
  contacts: [],
  contactDetails: { isLoading: false, data: {} },
  error: false,
};

const contactsList = [];

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
