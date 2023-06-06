import { FETCH_GROUPS_FULFILLED, FETCH_GROUPS_REJECTED } from '../types';

const initialState = {
  data: []
};

export default function (state = initialState, action) {
  switch (action.type) {
    case FETCH_GROUPS_FULFILLED:
      return { ...initialState, data: action.payload };
    case FETCH_GROUPS_REJECTED:
      return { ...initialState, data: [] };
    default:
      return state;
  }
}
