import { TOGGLE_FA_HEADER, TOGGLE_SIDEBAR_COLLAPSED_STATE } from './types';

export const toggleSidebarCollapsedStateAction = (payload) => ({
  type: TOGGLE_SIDEBAR_COLLAPSED_STATE,
  payload
});

export const toggleFaHeader = (payload) => ({
  type: TOGGLE_FA_HEADER,
  payload
});
