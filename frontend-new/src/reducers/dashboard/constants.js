import { EMPTY_OBJECT, EMPTY_ARRAY } from '../../utils/global';

export const defaultState = {
  dashboards: {
    loading: false,
    error: false,
    data: EMPTY_ARRAY
  },
  activeDashboard: EMPTY_OBJECT,
  activeDashboardUnits: {
    loading: false,
    error: false,
    data: EMPTY_ARRAY
  }
};
