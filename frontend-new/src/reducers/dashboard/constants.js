import { EMPTY_OBJECT, EMPTY_ARRAY } from '../../utils/global';

export const apiStates = {
  loading: false,
  error: false,
  completed: false
};

export const defaultState = {
  dashboards: {
    ...apiStates,
    data: EMPTY_ARRAY
  },
  activeDashboard: EMPTY_OBJECT,
  activeDashboardUnits: {
    ...apiStates,
    data: EMPTY_ARRAY
  },
  draftsSelected: false,
  showNewFolderModal: false,
  newFolderCreationState: {
    ...apiStates
  },
  addToExistingFolderState: {
    ...apiStates
  },
  renameFolderState: {
    ...apiStates
  },
  deleteFolderState: {
    ...apiStates
  },
  foldersList: {
    ...apiStates,
    data: EMPTY_ARRAY
  },
  dashboardsArrangementInFoldersCompleted: false,
  allBoardsFolderId: null,
  dashboardDeletionInitiated: false,
  editDashboardDetails: {
    initiated: false,
    editDashboard: null
  }
};
