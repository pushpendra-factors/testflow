/* eslint-disable */


import {
  FETCH_EVENTS,
  FETCH_EVENT_PROPERTIES,
  FETCH_USER_PROPERTIES,
  SET_GROUPBY,
  DEL_GROUPBY,
  INITIALIZE_GROUPBY,
  DEL_GROUPBY_EVENT
} from './actions';
import { SHOW_ANALYTICS_RESULT } from '../types';

const DEFAULT_TOUCHPOINTS = ['Campaign', 'Source', 'AdGroup', 'Keyword']

const defaultState = {
  eventOptions: [],
  eventProperties: {},
  userProperties: [],
  groupBy: {
    global: [],
    event: []
  },
  touchpointOptions: [{
    label: 'Paid Marketing',
    icon: 'fav',
    values: DEFAULT_TOUCHPOINTS.map(v => [v])
  }],
  show_analytics_result: false,
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
      let groupByState = Object.assign({}, state.groupBy);
      let gbp;
      if(groupByState[action.groupByType] && groupByState[action.groupByType][action.index]) {
        let groupTypeState = [...groupByState[action.groupByType]];
        if(groupTypeState[action.index] === action.payload) {
          groupTypeState.splice(action.index, 1);
          // groupTypeState.length -= 1;
        } else {
          gbp = groupTypeState.findIndex(i => i === state.payload)
          gbp && groupTypeState.splice(gbp, 1);
        }
        groupByState[action.groupByType] = groupTypeState;
        
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
    case DEL_GROUPBY_EVENT: {
      let groupByState = Object.assign({}, state.groupBy);
      let eventGroups = groupByState.event;
      const filteredEventGroups = eventGroups.filter(gbp => {
        return gbp.eventIndex !== action.index +1 && gbp.eventName !== action.payload.label
      })
      groupByState.event = filteredEventGroups;
      return { ...state, groupBy: groupByState };
    }
    case SHOW_ANALYTICS_RESULT: {
      return {
        ...state,
        show_analytics_result: action.payload,
      };
    }
    default:
      return state;
  }
}







