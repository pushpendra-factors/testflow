import { ACTIVE_DASHBOARD_CHANGE } from 'Reducers/types';

export const changeActiveDashboardAction = (newActiveDashboard) => {
  return { type: ACTIVE_DASHBOARD_CHANGE, payload: newActiveDashboard };
};
