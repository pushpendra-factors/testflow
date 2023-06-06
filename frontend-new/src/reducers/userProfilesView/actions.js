import {
  SET_TIMELINE_PAYLOAD,
  SET_PROFILES_ACTIVE_SEGMENT,
  UPDATE_TIMELINE_PAYLOAD,
  SET_PROFILE_SEGMENT_MODAL
} from './types';

export const setTimelinePayloadAction = (payload) => {
  return { type: SET_TIMELINE_PAYLOAD, payload };
};

export const setActiveSegmentAction = (payload) => {
  return { type: SET_PROFILES_ACTIVE_SEGMENT, payload };
};

export const updateTimelinePayloadAction = (payload) => {
  return { type: UPDATE_TIMELINE_PAYLOAD, payload };
};

export const setSegmentModalStateAction = (payload) => {
  return { type: SET_PROFILE_SEGMENT_MODAL, payload };
};
