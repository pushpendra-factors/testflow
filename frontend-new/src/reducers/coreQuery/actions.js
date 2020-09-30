/* eslint-disable */

export const FETCH_EVENTS = 'FETCH_EVENTS';

// Action creators
export const fetchEventsAction = (events, status = 'started') => {
  return { type: FETCH_EVENTS, payload: events };
};
