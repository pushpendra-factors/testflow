import {
  FETCH_GROUPS_FULFILLED,
  FETCH_GROUPS_REJECTED,
  SET_ACTIVE_PROJECT
} from './types';

const initialState = {
  data: {}
};

export default function (state = initialState, action) {
  switch (action.type) {
    case FETCH_GROUPS_FULFILLED:
      return { ...initialState, data: action.payload };
    case FETCH_GROUPS_REJECTED:
      return { ...initialState, data: {} };
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}
