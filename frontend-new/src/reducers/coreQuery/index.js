/* eslint-disable */


import {FETCH_EVENTS} from './actions';

const defaultState = {
  eventOptions: []
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_EVENTS:
      return { ...state, eventOptions: action.payload };
    default:
      return state;
  }
}







