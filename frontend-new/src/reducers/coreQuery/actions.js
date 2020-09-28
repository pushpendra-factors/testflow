/* eslint-disable */

export const FETCH_EVENTS = 'FETCH_EVENTS';

// Action creators
export const fetchEventsAction = (events, status = 'success') => {
  return { type: FETCH_EVENTS, payload: events };
};
