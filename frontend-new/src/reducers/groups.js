import { FETCH_GROUPS_FULFILLED, FETCH_GROUPS_REJECTED } from './types';

const inititalState = {
  data: []
};

export default function (state = inititalState, action) {
  switch (action.type) {
    case FETCH_GROUPS_FULFILLED:
      return { ...inititalState, data: action.payload };
    case FETCH_GROUPS_REJECTED:
      return { ...inititalState, data: [] };
    default:
      return state;
  }
}
