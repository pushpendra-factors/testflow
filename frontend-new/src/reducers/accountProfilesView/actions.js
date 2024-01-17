import {
  SET_ACCOUNT_PAYLOAD,
  UPDATE_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE,
  SET_FILTERS_DIRTY
} from './types';

export const setAccountPayloadAction = (payload) => ({
  type: SET_ACCOUNT_PAYLOAD,
  payload
});

export const updateAccountPayloadAction = (payload) => ({
  type: UPDATE_ACCOUNT_PAYLOAD,
  payload
});

export const setSegmentModalStateAction = (payload) => ({
  type: SET_ACCOUNT_SEGMENT_MODAL,
  payload
});

export const setNewSegmentModeAction = (payload) => {
  if (payload) {
    return { type: ENABLE_NEW_SEGMENT_MODE };
  }
  return { type: DISABLE_NEW_SEGMENT_MODE };
};

export const setFiltersDirtyAction = (payload) => ({
  type: SET_FILTERS_DIRTY,
  payload
});
