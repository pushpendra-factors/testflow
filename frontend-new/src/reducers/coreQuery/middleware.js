/* eslint-disable */

import { fetchEventsAction, fetchEventPropertiesAction, 
  fetchUserPropertiesAction, 
  setGroupByAction, delGroupByAction} from './actions';
import { getEventNames, fetchEventProperties, fetchUserProperties } from './services';
import { convertToEventOptions, convertPropsToOptions } from './utils';

export const fetchEventNames = (projectId) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      getEventNames(dispatch, projectId)
        .then((response) => {
          const options = convertToEventOptions(response.data.event_names);
          resolve(dispatch(fetchEventsAction(options)));
        }).catch((err) => {
          resolve(dispatch(fetchEventsAction([])));
        });
    });
  };
};
 
export const getUserProperties = (projectId, queryType) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchUserProperties(projectId, queryType).then((response) => {
        const options = convertPropsToOptions(response.data);
        resolve(dispatch(fetchUserPropertiesAction(options)));
      }).catch((err) => {
        // resolve(dispatch(fetchEventPropertiesAction({})));
      })
    })
  }
}

export const getEventProperties = (projectId, eventName) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchEventProperties(projectId, eventName)
        .then((response) => {
          const options = convertPropsToOptions(response.data);
          resolve(dispatch(fetchEventPropertiesAction(options, eventName)));
        }).catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
}

export const setGroupBy = (type, groupBy, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setGroupByAction(type, groupBy, index)))
    })
  }
}

export const delGroupBy = (type, payload, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(delGroupByAction(type, payload, index)))
    })
  }
}

