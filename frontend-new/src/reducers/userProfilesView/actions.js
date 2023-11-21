import {
  SET_TIMELINE_PAYLOAD,
  SET_PROFILES_ACTIVE_SEGMENT,
  UPDATE_TIMELINE_PAYLOAD,
  SET_PROFILE_SEGMENT_MODAL,
  SET_PROFILE_FILTERS_DIRTY,
  ENABLE_PROFILE_NEW_SEGMENT_MODE,
  DISABLE_PROFILE_NEW_SEGMENT_MODE
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

export const setNewSegmentModeAction = (payload) => {
  if (payload) {
    return { type: ENABLE_PROFILE_NEW_SEGMENT_MODE };
  } else {
    return { type: DISABLE_PROFILE_NEW_SEGMENT_MODE };
  }
};

export const setFiltersDirtyAction = (payload) => {
  return { type: SET_PROFILE_FILTERS_DIRTY, payload };
};
