/* eslint-disable */


import {FETCH_EVENTS, FETCH_EVENT_PROPERTIES, FETCH_USER_PROPERTIES} from './actions';

const defaultState = {
  eventOptions: [],
  eventProperties: {},
  userProperties: []
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_EVENTS:
      return { ...state, eventOptions: action.payload };
    case FETCH_USER_PROPERTIES:
      return { ...state, eventProperties: action.payload };
    case FETCH_EVENT_PROPERTIES:
      const eventPropState = Object.assign({}, state.eventProperties);
      eventPropState[action.eventName] = action.payload;
      return { ...state, eventProperties: eventPropState };  
    default:
      return state;
  }
}







