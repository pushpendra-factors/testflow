import {
  QUERIES_LOADING, QUERIES_LOADING_FAILED, QUERIES_LOADED, QUERY_CREATED, QUERY_DELETED
} from './types';

const inititalState = {
  loading: false,
  error: false,
  data: []
};

export default function (state = inititalState, action) {
  switch (action.type) {
    case QUERIES_LOADING:
      return { ...state, loading: true };
    case QUERIES_LOADING_FAILED:
      return { ...inititalState, error: true };
    case QUERIES_LOADED:
      return { ...inititalState, data: action.payload };
    case QUERY_CREATED:
      return { ...inititalState, data: [...state.data, action.payload] };
    case QUERY_DELETED:
      var index = state.data.findIndex(d => d.id === action.payload);
      return { ...inititalState, data: [...state.data.slice(0, index), ...state.data.slice(index + 1)] };
    default:
      return state;
  }
}
