import { fetchProfileAccountDetails } from '.';
import { formattedResponse, formattedResponseData } from './utils';

export const getProfileAccountDetails = (projectId, id) => {
  return (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
    return new Promise((resolve, reject) => {
      fetchProfileAccountDetails(projectId, id)
        .then((response) => {
          const data = formattedResponseData(response.data);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
              payload: data,
            })
          );
        })
        .catch((err) => {
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
              payload: {},
            })
          );
        });
    });
  };
};
