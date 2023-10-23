import { createSelector } from 'reselect';

export const selectDashboardList = (state) => state.dashboard.dashboards.data.filter(fil => fil.class !== 'predefined');

export const selectActiveDashboard = (state) => state.dashboard.activeDashboard;

export const selectAreDraftsSelected = (state) => state.dashboard.draftsSelected;

export const selectDashboardListFilteredBySearchText = createSelector(
  selectDashboardList,
  (state, searchText) => searchText,
  (dashboards, searchText) => {
    return dashboards.filter((d) =>
      d.name.toLowerCase().includes(searchText.toLowerCase())
    );
  }
);

// for pre-defined dashboards
export const selectPreDashboardList = (state) => state.dashboard.dashboards.data.filter(fil => fil.class === 'predefined');

export const selectActivePreDashboard = (state) => state.preBuildDashboardConfig.activePreBuildDashboard;

export const selectPreDashboardListFilteredBySearchText = createSelector(
  selectPreDashboardList,
  (state, searchText) => searchText,
  (dashboards, searchText) => {
    return dashboards.filter((d) =>
      d.name.toLowerCase().includes(searchText.toLowerCase())
    );
  }
);