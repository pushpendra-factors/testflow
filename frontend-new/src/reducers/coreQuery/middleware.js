/* eslint-disable */

import { fetchEventsAction, fetchEventPropertiesAction } from './actions';
import { getEventNames, fetchEventProperties, fetchUserProperties } from './services';
import { convertToEventOptions } from './utils';

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

export const getEventProperties = (projectId, eventName) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchEventProperties(projectId, eventName)
        .then((response) => {
          resolve(dispatch(fetchEventPropertiesAction(response.data, eventName)));
        }).catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
}

