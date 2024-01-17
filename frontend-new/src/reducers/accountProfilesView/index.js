import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { SEGMENT_DELETED } from 'Reducers/timelines/types';
import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  UPDATE_ACCOUNT_PAYLOAD,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE,
  SET_FILTERS_DIRTY
} from './types';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';

export const INITIAL_ACCOUNT_PAYLOAD = {
  source: GROUP_NAME_DOMAINS
};

const initialState = {
  accountPayload: INITIAL_ACCOUNT_PAYLOAD,
  showSegmentModal: false,
  newSegmentMode: false,
  filtersDirty: false
};

export default function (state = initialState, action) {
  switch (action.type) {
    case SET_ACCOUNT_PAYLOAD:
      return {
        ...state,
        accountPayload: action.payload,
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
        accountPayload: {
          ...INITIAL_ACCOUNT_PAYLOAD
        }
      };
    }
    case SET_FILTERS_DIRTY: {
      return {
        ...state,
        filtersDirty: action.payload
      };
    }
    default:
      return state;
  }
}
