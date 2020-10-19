/* eslint-disable */

export const FETCH_EVENTS = 'FETCH_EVENTS';
export const FETCH_EVENT_PROPERTIES = 'FETCH_EVENT_PROPERTIES';
export const FETCH_USER_PROPERTIES = 'FETCH_USER_PROPERTIES';

// Action creators
export const fetchEventsAction = (events, status = 'started') => {
  return { type: FETCH_EVENTS, payload: events };
};

export const fetchUserPropertiesAction = (userProps) => {
  return { type: FETCH_USER_PROPERTIES, payload: userProps};
}

export const fetchEventPropertiesAction = (eventProps, name) => {
  return { type: FETCH_EVENT_PROPERTIES, payload: eventProps, eventName: name };
}
