import {
  SET_TIMELINE_PAYLOAD,
  UPDATE_TIMELINE_PAYLOAD,
  SET_PROFILE_SEGMENT_MODAL,
  SET_PROFILE_FILTERS_DIRTY,
  ENABLE_PROFILE_NEW_SEGMENT_MODE,
  DISABLE_PROFILE_NEW_SEGMENT_MODE
} from './types';

export const setTimelinePayloadAction = (payload) => ({
  type: SET_TIMELINE_PAYLOAD,
  payload
});

export const updateTimelinePayloadAction = (payload) => ({
  type: UPDATE_TIMELINE_PAYLOAD,
  payload
});

export const setSegmentModalStateAction = (payload) => ({
  type: SET_PROFILE_SEGMENT_MODAL,
  payload
});

export const setNewSegmentModeAction = (payload) => {
  if (payload) {
    return { type: ENABLE_PROFILE_NEW_SEGMENT_MODE };
  }
  return { type: DISABLE_PROFILE_NEW_SEGMENT_MODE };
};

export const setFiltersDirtyAction = (payload) => ({
  type: SET_PROFILE_FILTERS_DIRTY,
  payload
});
