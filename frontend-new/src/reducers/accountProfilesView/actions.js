import {
  SET_ACCOUNT_PAYLOAD,
  UPDATE_ACCOUNT_PAYLOAD,
  SET_ACCOUNT_SEGMENT_MODAL,
  ENABLE_NEW_SEGMENT_MODE,
  DISABLE_NEW_SEGMENT_MODE,
  SET_FILTERS_DIRTY,
  TOGGLE_ACCOUNTS_TAB,
  SET_INSIGHTS_DURATION,
  SET_INSIGHTS_COMPARE_SEGMENT,
  RESET_EDIT_INSIGHTS_METRIC
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

export const toggleAccountsTab = (payload) => ({
  type: TOGGLE_ACCOUNTS_TAB,
  payload
});

export const setInsightsDuration = (payload) => ({
  type: SET_INSIGHTS_DURATION,
  payload
});

export const setInsightsCompareSegment = (segmentId, compareSegmentId) => ({
  type: SET_INSIGHTS_COMPARE_SEGMENT,
  payload: { segmentId, compareSegmentId }
});

export const setDrawerVisibleAction = (isVisible) => ({
  type: 'SET_DRAWER_VISIBLE',
  payload: isVisible
});

export const setActiveDomainAction = (domainData) => ({
  type: 'SET_ACTIVE_DOMAIN',
  payload: domainData
});

export const resetEditMetricStatus = () => ({
  type: RESET_EDIT_INSIGHTS_METRIC
});
