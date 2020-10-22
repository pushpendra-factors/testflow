/* eslint-disable */


import {FETCH_EVENTS, 
  FETCH_EVENT_PROPERTIES, 
  FETCH_USER_PROPERTIES, 
  SET_GROUPBY,
} from './actions';

const defaultState = {
  eventOptions: [],
  eventProperties: {},
  userProperties: [],
  groupBy: {
    global: [],
    event: []
  }
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_EVENTS:
      return { ...state, eventOptions: action.payload };
    case FETCH_USER_PROPERTIES:
      return { ...state, userProperties: action.payload };
    case FETCH_EVENT_PROPERTIES:
      const eventPropState = Object.assign({}, state.eventProperties);
      eventPropState[action.eventName] = action.payload;
      return { ...state, eventProperties: eventPropState };
    case SET_GROUPBY:
      const groupByState = Object.assign({}, state.groupBy);
      if(groupByState[action.groupByType] && groupByState[action.groupByType][action.index]) {
        groupByState[action.groupByType][action.index] = action.payload;
      } else if (groupByState[action.groupByType] && action.index === groupByState[action.groupByType].length) {
        groupByState[action.groupByType].push(action.payload);
      }
      return { ...state, groupBy: groupByState };
    default:
      return state;
  }
}







