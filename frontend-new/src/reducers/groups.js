import {
  GROUPS_LOADING,
  GROUPS_LOADING_FAILED,
  GROUPS_LOADED,
} from './types';

const inititalState = {
  loading: false,
  error: false,
  data: [],
};

export default function (state = inititalState, action) {
  switch (action.type) {
    case GROUPS_LOADING:
      return { ...state, loading: true };
    case GROUPS_LOADING_FAILED:
      return { ...inititalState, error: true };
    case GROUPS_LOADED:
      return { ...inititalState, data: action.payload };
    default:
      return state;
  }
}
