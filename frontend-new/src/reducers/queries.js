import {
  QUERIES_LOADING,
  QUERIES_LOADING_FAILED,
  QUERIES_LOADED,
  QUERY_CREATED,
  QUERY_DELETED,
  QUERIES_LOADING_STOPPED,
  QUERY_UPDATED,
} from './types';

const inititalState = {
  loading: false,
  error: false,
  data: [],
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
      return { ...inititalState, data: [action.payload, ...state.data] };
    case QUERY_DELETED:
      var index = state.data.findIndex((d) => d.id === action.payload);
      return {
        ...inititalState,
        data: [...state.data.slice(0, index), ...state.data.slice(index + 1)],
      };
    case QUERIES_LOADING_STOPPED:
      return { ...state, loading: false };
    case QUERY_UPDATED:
      const updatedQueryIndex = state.data.findIndex(
        (d) => d.id === action.queryId
      );
      if (updatedQueryIndex > -1) {
        return {
          ...state,
          data: [
            ...state.data.slice(0, updatedQueryIndex),
            { ...state.data[updatedQueryIndex], ...action.payload },
            ...state.data.slice(updatedQueryIndex + 1),
          ],
        };
      }
      return state;
    default:
      return state;
  }
}
