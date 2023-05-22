import { createSelector } from 'reselect';

export const selectDashboardList = (state) => state.dashboard.dashboards.data;

export const selectActiveDashboard = (state) => state.dashboard.activeDashboard;

export const selectDashboardListFilteredBySearchText = createSelector(
  selectDashboardList,
  (state, searchText) => searchText,
  (dashboards, searchText) => {
    return dashboards.filter((d) =>
      d.name.toLowerCase().includes(searchText.toLowerCase())
    );
  }
);
