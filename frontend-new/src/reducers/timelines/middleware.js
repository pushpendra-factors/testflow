import { fetchProfileAccounts, fetchProfileAccountDetails } from '.';
import { formattedResponseData } from './utils';

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

export const getProfileAccountDetails = (projectId, id) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
  return new Promise((resolve) => {
    fetchProfileAccountDetails(projectId, id)
      .then((response) => {
        const data = formattedResponseData(response.data);
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
