import { createSelector } from 'reselect';

export const selectDashboardList = (state) => state.dashboard.dashboards.data;

export const selectActiveDashboard = (state) => state.dashboard.activeDashboard;

export const selectAreDraftsSelected = (state) =>
  state.dashboard.draftsSelected;

export const selectDashboardListFilteredBySearchText = createSelector(
  selectDashboardList,
  (state, searchText) => searchText,
  (dashboards, searchText) =>
    dashboards.filter((d) =>
      d.name.toLowerCase().includes(searchText.toLowerCase())
    )
);

// for pre-defined dashboards
export const selectActivePreDashboard = (state) =>
  state.preBuildDashboardConfig.activePreBuildDashboard;

export const selectShowDashboardNewFolderModal = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.showNewFolderModal
);

export const selectNewFolderCreationState = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.newFolderCreationState
);

export const selectDashboardFoldersListState = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.foldersList
);

export const selectAllBoardsFolderId = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.allBoardsFolderId
);

export const selectRenameFolderState = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.renameFolderState
);

export const selectDeleteFolderState = createSelector(
  (state) => state.dashboard,
  (dashboardState) => dashboardState.deleteFolderState
);
