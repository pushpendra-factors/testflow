import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACTIVE_SEGMENT,
  UPDATE_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE,
  SET_FILTERS_DIRTY,
  SET_EXIT_CONFIRMATION_MODAL
} from './types';

export const setAccountPayloadAction = (payload) => {
  return { type: SET_ACCOUNT_PAYLOAD, payload };
};

export const setActiveSegmentAction = (payload) => {
  return { type: SET_ACTIVE_SEGMENT, payload };
};

export const updateAccountPayloadAction = (payload) => {
  return { type: UPDATE_ACCOUNT_PAYLOAD, payload };
};

export const setSegmentModalStateAction = (payload) => {
  return { type: SET_ACCOUNT_SEGMENT_MODAL, payload };
};

export const setNewSegmentModeAction = (payload) => {
  if (payload) {
    return { type: ENABLE_NEW_SEGMENT_MODE };
  } else {
    return { type: DISABLE_NEW_SEGMENT_MODE };
  }
};

export const setFiltersDirtyAction = (payload) => {
  return { type: SET_FILTERS_DIRTY, payload };
};

export const setExitConfirmationModalAction = (value, routingCallback) => {
  return {
    type: SET_EXIT_CONFIRMATION_MODAL,
    payload: { value, routingCallback }
  };
};
