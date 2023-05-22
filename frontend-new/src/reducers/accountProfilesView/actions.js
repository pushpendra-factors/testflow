import {
  SET_ACCOUNT_PAYLOAD,
  SET_ACTIVE_SEGMENT,
  UPDATE_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL
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
