import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  SET_ACTIVE_SEGMENT,
  UPDATE_ACCOUNT_PAYLOAD,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE
} from './types';
import { SEGMENT_DELETED } from 'Reducers/timelines/types';

const INITIAL_ACTIVE_SEGMENT = {};
const INITIAL_ACCOUNT_PAYLOAD = { source: '', filters: [], segment_id: '' };

const initialState = {
  accountPayload: INITIAL_ACCOUNT_PAYLOAD,
  activeSegment: INITIAL_ACTIVE_SEGMENT,
  showSegmentModal: false,
  newSegmentMode: false
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: action.payload,
        newSegmentMode: false
      };
    case SET_ACTIVE_SEGMENT:
      return {
        ...state,
        activeSegment: action.payload,
        newSegmentMode: false
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
    case ENABLE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: true
      };
    }
    case DISABLE_NEW_SEGMENT_MODE: {
      return {
        ...state,
        newSegmentMode: false
      };
    }
    case SET_ACTIVE_PROJECT: {
      return {
        ...initialState
      };
    }
    case SEGMENT_DELETED: {
      return {
        ...state,
        activeSegment: INITIAL_ACTIVE_SEGMENT,
        accountPayload: { ...INITIAL_ACCOUNT_PAYLOAD, source: 'All' }
      };
    }
    default:
      return state;
  }
}
