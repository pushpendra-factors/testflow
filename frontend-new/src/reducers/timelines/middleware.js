import logger from 'Utils/logger';
import {
  fetchProfileAccounts,
  fetchProfileAccountDetails,
  fetchProfileUsers,
  fetchProfileUserDetails,
  createSegment,
  fetchSegments,
  updateSegment,
  deleteSegmentByID,
  fetchAccountOverview,
  fetchTop100Events,
  fetchConfiguredUserProperties,
  fetchConfiguredEventProperties,
  fetchSegmentFolders
} from '.';
import { formatAccountTimeline, formatUsersTimeline } from './utils';
import {
  deleteSegmentAction,
  setLoadingSegmentFolders,
  setSegmentFolders
} from './actions';
import {
  SET_ACCOUNTS_SEGMENT_FOLDERS_FAILED,
  SET_PEOPLES_SEGMENT_FOLDERS_FAILED
} from './types';

export const getProfileAccounts = (projectId, payload) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_ACCOUNTS_LOADING' });
  return new Promise((resolve, reject) => {
    fetchProfileAccounts(projectId, payload)
      .then((response) => {
        const data = response.data?.map((account) => ({
          ...account,
          domain: { id: account.identity, name: account?.domain_name }
        }));
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNTS_FULFILLED',
            payload: data,
            segmentID: payload.segment_id || 'default',
            status: response.status
          })
        );
      })
      .catch((err) => {
        dispatch({
          type: 'FETCH_PROFILE_ACCOUNTS_FAILED',
          segmentID: payload.segment_id || 'default'
        });
        reject(err);
      });
  });
};

export const getProfileAccountDetails =
  (projectId, id, source) => (dispatch) => {
    dispatch({ type: 'FETCH_PROFILE_ACCOUNT_DETAILS_LOADING' });
    return new Promise((resolve) => {
      fetchProfileAccountDetails(projectId, id, source)
        .then((response) => {
          const data = formatAccountTimeline(response.data);
          const userPropertiesMap = {};
          const eventPropertiesMap = {};
          data.events.forEach((event) => {
            eventPropertiesMap[event.id] = event.properties;
          });
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FULFILLED',
            payload: data
          });
          dispatch({
            type: 'FETCH_USER_CONFIG_PROPERTIES_MAP_FULFILLED',
            payload: userPropertiesMap
          });
          dispatch({
            type: 'FETCH_EVENT_CONFIG_PROPERTIES_MAP_FULFILLED',
            payload: eventPropertiesMap
          });
        })
        .catch((err) => {
          logger.error(err);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_ACCOUNT_DETAILS_FAILED',
              payload: {}
            })
          );
        });
    });
  };

export const getAccountOverview = (projectId, source, id) => (dispatch) => {
  dispatch({ type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_LOADING' });
  return new Promise((resolve) => {
    fetchAccountOverview(projectId, source, id)
      .then((response) => {
        const data = { ...response.data, id };
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FULFILLED',
            payload: data
          })
        );
      })
      .catch((err) => {
        logger.error(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_ACCOUNT_OVERVIEW_FAILED',
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
        const data = response.data?.map((user) => ({
          ...user,
          identity: { id: user.identity, isAnonymous: user.is_anonymous },
          tableProps: user.table_props,
          lastActivity: user.last_activity
        }));
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FULFILLED',
            payload: data,
            status: response.status
          })
        );
      })
      .catch((err) => {
        logger.error(err);
        resolve(
          dispatch({
            type: 'FETCH_PROFILE_USERS_FAILED',
            payload: [],
            error: err
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
          logger.error(err);
          resolve(
            dispatch({
              type: 'FETCH_PROFILE_USER_DETAILS_FAILED',
              payload: {}
            })
          );
        });
    });
  };

export const createNewSegment = (projectId, payload) => (dispatch) =>
  new Promise((resolve, reject) => {
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

export const getSavedSegments = (projectId) => (dispatch) =>
  new Promise((resolve, reject) => {
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

export const updateSegmentForId = (projectId, id, payload) => (dispatch) =>
  new Promise((resolve, reject) => {
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

export const deleteSegment =
  ({ projectId, segmentId }) =>
  (dispatch) =>
    new Promise((resolve, reject) => {
      deleteSegmentByID({ projectId, segmentId })
        .then(() => {
          dispatch(deleteSegmentAction({ segmentId }));
          resolve();
        })
        .catch((err) => {
          reject(err);
        });
    });

export const setActivePageviewEvent = (eventName) => ({
  type: 'SET_PAGEVIEW',
  payload: eventName
});

export const getTop100Events = (projectID, domainName) => async (dispatch) => {
  dispatch({ type: 'FETCH_TOP100_EVENTS_LOADING', domainName });
  try {
    const response = await fetchTop100Events(projectID, domainName);
    const events = response.data.map((event) => ({
      ...event,
      username: event.username || event.user_id,
      enabled: true
    }));
    const eventPropertiesMap = {};
    events.forEach((event) => {
      eventPropertiesMap[event.id] = event.properties;
    });
    dispatch({
      type: 'FETCH_TOP100_EVENTS_FULFILLED',
      payload: events || [],
      domainName
    });
    dispatch({
      type: 'FETCH_EVENT_CONFIG_PROPERTIES_MAP_FULFILLED',
      payload: eventPropertiesMap
    });
  } catch (err) {
    dispatch({ type: 'FETCH_TOP100_EVENTS_FAILED', domainName });
    logger.error(err);
  }
};

export const getConfiguredUserProperties =
  (projectID, userID, isAnonymous) => (dispatch) =>
    new Promise((resolve, reject) => {
      fetchConfiguredUserProperties(projectID, userID, isAnonymous)
        .then((response) => {
          resolve(
            dispatch({
              type: 'FETCH_USER_CONFIG_PROPERTIES_FULFILLED',
              payload: response.data,
              userID
            })
          );
        })
        .catch((err) => {
          reject(err);
        });
    });

export const getConfiguredEventProperties =
  (projectID, eventID, eventName) => (dispatch) =>
    new Promise((resolve, reject) => {
      fetchConfiguredEventProperties(projectID, eventID, eventName)
        .then((response) => {
          resolve(
            dispatch({
              type: 'FETCH_EVENT_CONFIG_PROPERTIES_FULFILLED',
              payload: response.data,
              eventID
            })
          );
        })
        .catch((err) => {
          reject(err);
        });
    });

export const getSegmentFolders =
  (projectId, folder_type = 'account') =>
  async (dispatch) => {
    setLoadingSegmentFolders();

    fetchSegmentFolders(projectId, folder_type)
      .then((response) => {
        dispatch(setSegmentFolders({ folder_type, data: response.data }));
      })
      .catch(() => {
        if (folder_type === 'account')
          dispatch({ type: SET_ACCOUNTS_SEGMENT_FOLDERS_FAILED });
        else dispatch({ type: SET_PEOPLES_SEGMENT_FOLDERS_FAILED });
      });
  };
