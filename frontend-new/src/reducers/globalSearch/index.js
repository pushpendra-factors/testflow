import { defaultState } from './constants';

const { TOGGLE_GLOBAL_SEARCH, SET_ACTIVE_PROJECT } = require('Reducers/types');

export default function (state = defaultState, action) {
  switch (action.type) {
    case TOGGLE_GLOBAL_SEARCH:
      return { ...state, visible: !state.visible };
    case SET_ACTIVE_PROJECT: {
      return {
        ...defaultState
      };
    }
    default:
      return state;
  }
}
