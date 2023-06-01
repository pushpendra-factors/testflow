import { defaultState } from './constants';

const { UPDATE_ALL_ROUTES, SET_ACTIVE_PROJECT } = require('Reducers/types');

export default function (state = defaultState, action) {
  switch (action.type) {
    case UPDATE_ALL_ROUTES:
      return { ...state, data: new Set([...state.data, ...action.payload]) };
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };
    default:
      return state;
  }
}
