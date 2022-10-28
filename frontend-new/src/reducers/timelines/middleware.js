import {
  fetchProfileAccounts,
  fetchProfileAccountDetails,
  fetchProfileUsers,
  fetchProfileUserDetails
} from '.';
import { formatAccountTimeline, formatUsersTimeline } from './utils';

export const getProfileAccounts = (projectId, payload) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_ACCOUNTS_LOADING' });
  return new Promise((resolve) => {
    fetchProfileAccounts(projectId, payload)
      .then((response) => {
        const data = response.data.map((account) => ({
          identity: account.identity,
          account: { name: account.name, host: account?.host_name },
          associated_contacts: account?.associated_contacts,
          country: account.country,
          last_activity: account.last_activity
        }));
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNTS_FULFILLED',
            payload: data
          })
        );
      })
      .catch((err) => {
        console.log(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNTS_FULFILLED',
            payload: {}
          })
        );
      });
  });
};

export const getProfileAccountDetails =
  (projectId, id, config) => (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
    return new Promise((resolve) => {
      fetchProfileAccountDetails(projectId, id)
        .then((response) => {
          const data = formatAccountTimeline(response.data, config);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
              payload: data
            })
          );
        })
        .catch((err) => {
          console.log(err);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
              payload: {}
            })
          );
        });
    });
  };

export const getProfileUsers = (projectId, payload) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_USERS_LOADING' });
  return new Promise((resolve) => {
    fetchProfileUsers(projectId, payload)
      .then((response) => {
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FULFILLED',
            payload: response.data
          })
        );
      })
      .catch((err) => {
        console.log(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FULFILLED',
            payload: {}
          })
        );
      });
  });
};

export const getProfileUserDetails =
  (projectId, id, isAnonymous, config) => (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_USER_DETAILS_LOADING' });
    return new Promise((resolve) => {
      fetchProfileUserDetails(projectId, id, isAnonymous)
        .then((response) => {
          const data = formatUsersTimeline(response.data, config);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_USER_DETAILS_FULFILLED',
              payload: data
            })
          );
        })
        .catch((err) => {
          console.log(err);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_USER_DETAILS_FULFILLED',
              payload: {}
            })
          );
        });
    });
  };
