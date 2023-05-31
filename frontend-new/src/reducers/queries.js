import {
  QUERIES_LOADING,
  QUERIES_LOADING_FAILED,
  QUERIES_LOADED,
  QUERY_CREATED,
  QUERY_DELETED,
  QUERIES_LOADING_STOPPED,
  QUERY_UPDATED,
  SET_ACTIVE_PROJECT
} from './types';

const initialState = {
  loading: false,
  error: false,
  data: []
};

export default function (state = initialState, action) {
  switch (action.type) {
    case QUERIES_LOADING:
      return { ...state, loading: true };
    case QUERIES_LOADING_FAILED:
      return { ...initialState, error: true };
    case QUERIES_LOADED:
      return { ...initialState, data: action.payload };
    case QUERY_CREATED:
      return { ...initialState, data: [action.payload, ...state.data] };
    case QUERY_DELETED:
      var index = state.data.findIndex((d) => d.id === action.payload);
      return {
        ...initialState,
        data: [...state.data.slice(0, index), ...state.data.slice(index + 1)]
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
            ...state.data.slice(updatedQueryIndex + 1)
          ]
        };
      }
      return state;
    case SET_ACTIVE_PROJECT:
      return {
        ...initialState
      };
    default:
      return state;
  }
}
