/* eslint-disable */

import { fetchEventsAction } from './actions';
import { getEventNames } from './services';
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
