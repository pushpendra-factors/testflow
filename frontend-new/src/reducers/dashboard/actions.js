import { ACTIVE_DASHBOARD_CHANGE } from 'Reducers/types';
import { SET_DRAFTS_SELECTED } from './types';

export const changeActiveDashboardAction = (newActiveDashboard) => {
  return { type: ACTIVE_DASHBOARD_CHANGE, payload: newActiveDashboard };
};

export const makeDraftsActiveAction = () => {
  return {
    type: SET_DRAFTS_SELECTED
  };
};
