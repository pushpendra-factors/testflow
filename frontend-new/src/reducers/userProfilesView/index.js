import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import {
  SET_TIMELINE_PAYLOAD,
  UPDATE_TIMELINE_PAYLOAD,
  SET_PROFILE_SEGMENT_MODAL,
  SET_PROFILE_FILTERS_DIRTY,
  ENABLE_PROFILE_NEW_SEGMENT_MODE,
  DISABLE_PROFILE_NEW_SEGMENT_MODE
} from './types';

const INITIAL_TIMELINE_PAYLOAD = { source: 'All', segment: {} };

const initialState = {
  timelinePayload: INITIAL_TIMELINE_PAYLOAD,
  showSegmentModal: false,
  newSegmentMode: false,
  filtersDirty: false
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_TIMELINE_PAYLOAD:
      return {
        ...state,
        timelinePayload: action.payload,
        newSegmentMode: false
      };
    case UPDATE_TIMELINE_PAYLOAD:
      return {
        ...state,
        timelinePayload: {
          ...state.timelinePayload,
          ...action.payload
        }
      };
    case SET_PROFILE_SEGMENT_MODAL: {
      return {
        ...state,
        showSegmentModal: action.payload
      };
    }
    case SET_ACTIVE_PROJECT: {
      return {
        ...initialState
      };
    }
    case SET_PROFILE_FILTERS_DIRTY: {
      return {
        ...state,
        filtersDirty: action.payload
      };
    }
    case ENABLE_PROFILE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: true
      };
    }
    case DISABLE_PROFILE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: false
      };
    }
    default:
      return state;
  }
}
