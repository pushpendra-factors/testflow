import {
  fetchProfileAccounts,
  fetchProfileAccountDetails,
  fetchProfileUsers,
  fetchProfileUserDetails,
  createSegment,
  fetchSegments,
  updateSegment,
  deleteSegmentByID,
  fetchAccountOverview
} from '.';
import { formatAccountTimeline, formatUsersTimeline } from './utils';
import { deleteSegmentAction } from './actions';

export const getProfileAccounts =
  (projectId, payload, agentId) => (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_ACCOUNTS_LOADING' });
    return new Promise((resolve) => {
      fetchProfileAccounts(projectId, payload, agentId)
        .then((response) => {
          const data = response.data.map((account) => ({
            ...account,
            identity: account.identity,
            account: { name: account.name, host: account?.host_name },
            tableProps: account.table_props,
            lastActivity: account.last_activity
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
              type: 'FETCH_PROFILE_ACCOUNTS_FAILED',
              payload: []
            })
          );
        });
    });
  };

export const getProfileAccountDetails =
  (projectId, id, source, config) => (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
    return new Promise((resolve) => {
      fetchProfileAccountDetails(projectId, id, source)
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
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED',
              payload: {}
            })
          );
        });
    });
  };

export const getAccountOverview = (projectId, id, source) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_LOADING' });
  return new Promise((resolve) => {
    fetchAccountOverview(projectId, id, source)
      .then((response) => {
        const data = { ...response.data, id: id };
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FULFILLED',
            payload: data
          })
        );
      })
      .catch((err) => {
        console.log(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FAILED',
            payload: {}
          })
        );
      });
  });
};

export const getProfileUsers = (projectId, payload, agentId) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_USERS_LOADING' });
  return new Promise((resolve) => {
    fetchProfileUsers(projectId, payload, agentId)
      .then((response) => {
        const data = response.data.map((user) => ({
          ...user,
          identity: { id: user.identity, isAnonymous: user.is_anonymous },
          tableProps: user.table_props,
          lastActivity: user.last_activity
        }));
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FULFILLED',
            payload: data
          })
        );
      })
      .catch((err) => {
        console.log(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FAILED',
            payload: []
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
              type: 'FETCH_PROFILE_USER_DETAILS_FAILED',
              payload: {}
            })
          );
        });
    });
  };

export const createNewSegment = (projectId, payload) => (dispatch) => {
  return new Promise((resolve, reject) => {
    createSegment(projectId, payload)
      .then((response) => {
        resolve(
          dispatch({
            type: 'SEGMENT_CREATION_FULFILLED',
            payload: response.data
          })
        );
      })
      .catch((err) => {
        dispatch({ type: 'SEGMENT_CREATION_REJECTED', payload: err });
        reject(err);
      });
  });
};

export const getSavedSegments = (projectId) => (dispatch) => {
  return new Promise((resolve, reject) => {
    fetchSegments(projectId)
      .then((response) => {
        resolve(
          dispatch({
            type: 'FETCH_SEGMENTS_FULFILLED',
            payload: response.data
          })
        );
      })
      .catch((err) => {
        dispatch({ type: 'FETCH_SEGMENTS_REJECTED', payload: err });
        reject(err);
      });
  });
};

export const updateSegmentForId = (projectId, id, payload) => (dispatch) => {
  return new Promise((resolve, reject) => {
    updateSegment(projectId, id, payload)
      .then((response) => {
        resolve(
          dispatch({
            type: 'UPDATE_SEGMENT_FULFILLED',
            payload: response.data
          })
        );
      })
      .catch((err) => {
        dispatch({ type: 'UPDATE_SEGMENT_REJECTED', payload: err });
        reject(err);
      });
  });
};

export const deleteSegment = ({ projectId, segmentId }) => {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      deleteSegmentByID({ projectId, segmentId })
        .then((response) => {
          dispatch(deleteSegmentAction({ segmentId }));
          resolve();
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
};
