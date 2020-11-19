/* eslint-disable */


import {
  FETCH_EVENTS,
  FETCH_EVENT_PROPERTIES,
  FETCH_USER_PROPERTIES,
  SET_GROUPBY,
  DEL_GROUPBY,
  INITIALIZE_GROUPBY
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
    case INITIALIZE_GROUPBY: {
      return {
        ...state, groupBy: action.payload
      }
    }
    case DEL_GROUPBY: {
      const groupByState = Object.assign({}, state.groupBy);
      let gbp;
      if(groupByState[action.groupByType] && groupByState[action.groupByType][action.index]) {
        if(groupByState[action.groupByType][action.index] === action.payload) {
          delete groupByState[action.groupByType][action.index]
          groupByState[action.groupByType].length -= 1;
        } else {
          gbp = groupByState[action.groupByType].findIndex(i => i === state.payload)
          gbp && delete groupByState[gbp];
          groupByState[action.groupByType].length -= 1;
        }
        
      }
      return { ...state, groupBy: groupByState };
    }
      
    case SET_GROUPBY:
      let groupByState = Object.assign({}, state.groupBy);
      if (groupByState[action.groupByType] && groupByState[action.groupByType][action.index]) {
        groupByState[action.groupByType][action.index] = action.payload;
      } else if (groupByState[action.groupByType] && action.index === groupByState[action.groupByType].length) {
        groupByState[action.groupByType].push(action.payload);
      }
      groupByState[action.groupByType].sort((a, b) => {
        return a.prop_category >= b.prop_category ? 1 : -1;
      });
      return { ...state, groupBy: groupByState };
    default:
      return state;
  }
}







