import {
  SET_TIMELINE_PAYLOAD,
  SET_PROFILES_ACTIVE_SEGMENT,
  UPDATE_TIMELINE_PAYLOAD,
  SET_PROFILE_SEGMENT_MODAL
} from './types';

const INITIAL_ACTIVE_SEGMENT = {};
const INITIAL_TIMELINE_PAYLOAD = { source: 'web', filters: [], segment_id: '' };

const initialState = {
  timelinePayload: INITIAL_TIMELINE_PAYLOAD,
  activeSegment: INITIAL_ACTIVE_SEGMENT,
  showSegmentModal: false
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_TIMELINE_PAYLOAD:
      return {
        ...state,
        timelinePayload: action.payload,
        activeSegment: INITIAL_ACTIVE_SEGMENT
      };
    case SET_PROFILES_ACTIVE_SEGMENT:
      return {
        ...state,
        activeSegment: action.payload.segmentPayload,
        timelinePayload: action.payload.timelinePayload
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
    default:
      return state;
  }
}
