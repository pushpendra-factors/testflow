import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  SET_ACTIVE_SEGMENT,
  UPDATE_ACCOUNT_PAYLOAD
} from './types';

const INITIAL_ACTIVE_SEGMENT = {};
const INITIAL_ACCOUNT_PAYLOAD = { source: '', filters: [], segment_id: '' };

const initialState = {
  accountPayload: INITIAL_ACCOUNT_PAYLOAD,
  activeSegment: INITIAL_ACTIVE_SEGMENT,
  showSegmentModal: false
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: action.payload,
        activeSegment: INITIAL_ACTIVE_SEGMENT
      };
    case SET_ACTIVE_SEGMENT:
      return {
        ...state,
        activeSegment: action.payload.segmentPayload,
        accountPayload: action.payload.accountPayload
      };
    case UPDATE_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: {
          ...state.accountPayload,
          ...action.payload
        }
      };
    case SET_ACCOUNT_SEGMENT_MODAL: {
      return {
        ...state,
        showSegmentModal: action.payload
      };
    }
    default:
      return state;
  }
}
