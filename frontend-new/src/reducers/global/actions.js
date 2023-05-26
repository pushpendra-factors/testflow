import { TOGGLE_SIDEBAR_COLLAPSED_STATE } from './types';

export const toggleSidebarCollapsedStateAction = (payload) => {
  return {
    type: TOGGLE_SIDEBAR_COLLAPSED_STATE,
    payload
  };
};
