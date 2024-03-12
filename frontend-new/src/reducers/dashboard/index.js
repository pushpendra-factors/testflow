import _ from 'lodash';

import {
  DASHBOARDS_LOADED,
  DASHBOARDS_LOADING,
  DASHBOARDS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADING,
  DASHBOARD_UNITS_LOADING_FAILED,
  DASHBOARD_UNITS_LOADED,
  ACTIVE_DASHBOARD_CHANGE,
  DASHBOARD_CREATED,
  DASHBOARD_DELETED,
  UNITS_ORDER_CHANGED,
  DASHBOARD_UNMOUNTED,
  WIDGET_DELETED,
  DASHBOARD_UPDATED,
  SET_ACTIVE_PROJECT,
  DASHBOARD_LAST_REFRESHED,
  ADD_DASHBOARD_MODAL_CLOSE
} from '../types';

import {
  SET_DRAFTS_SELECTED,
  TOGGLE_DASHBOARD_NEW_FOLDER_MODAL,
  INITIATED_DASHBOARD_NEW_FOLDER_CREATION,
  DASHBOARD_NEW_FOLDER_CREATION_SUCCESSFUL,
  DASHBOARD_NEW_FOLDER_CREATION_FAILED,
  DASHBOARD_FOLDERS_LIST_LOADING,
  DASHBOARD_FOLDERS_LIST_SUCCESS,
  DASHBOARD_FOLDERS_LIST_ERROR,
  INITIATED_DASHBOARD_MOVE_TO_EXISTING_FOLDER,
  DASHBOARD_MOVE_TO_EXISTING_FOLDER_SUCCESSFUL,
  DASHBOARD_MOVE_TO_EXISTING_FOLDER_FAILED,
  INITIATED_RENAME_DASHBOARD_FOLDER,
  RENAME_DASHBOARD_FOLDER_SUCCESSFUL,
  RENAME_DASHBOARD_FOLDER_FAILED,
  INITIATED_DELETE_DASHBOARD_FOLDER,
  DELETE_DASHBOARD_FOLDER_SUCCESSFUL,
  DELETE_DASHBOARD_FOLDER_FAILED,
  INITIATE_DASHBOARD_DELETION,
  RESET_DASHBOARD_DELETION_INITIATION,
  INITIATE_EDIT_DASHBOARD_DETAILS
} from './types';

import {
  getRearrangedData,
  getUpdateStateOnDashboardsLoaded,
  getFoldersListWithDashboardIds,
  getAllBoardsFolderId
} from './utils';
import { defaultState, apiStates } from './constants';

export default function (state = defaultState, action) {
  switch (action.type) {
    case DASHBOARDS_LOADING:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, loading: true }
      };
    case DASHBOARDS_LOADING_FAILED:
      return {
        ...defaultState,
        dashboards: { ...defaultState.dashboards, error: true }
      };
    case DASHBOARDS_LOADED:
      if (
        state.foldersList.completed &&
        state.dashboardsArrangementInFoldersCompleted === false
      ) {
        const foldersListWithDashboardIds = getFoldersListWithDashboardIds(
          action.payload,
          state.foldersList.data
        );
        return {
          ...getUpdateStateOnDashboardsLoaded({
            payload: action.payload
          }),
          dashboardsArrangementInFoldersCompleted: true,
          foldersList: {
            ...state.foldersList,
            data: foldersListWithDashboardIds
          }
        };
      }
      return getUpdateStateOnDashboardsLoaded({
        payload: action.payload
      });
    case DASHBOARD_UNITS_LOADING:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          loading: true
        }
      };
    case DASHBOARD_UNITS_LOADING_FAILED:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          error: true
        }
      };
    case DASHBOARD_UNITS_LOADED:
      return {
        ...state,
        activeDashboardUnits: {
          ...defaultState.activeDashboardUnits,
          data: getRearrangedData(action.payload, state.activeDashboard)
        }
      };
    case ACTIVE_DASHBOARD_CHANGE:
      return {
        ...state,
        activeDashboard: action.payload,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
        draftsSelected: false
      };
    case DASHBOARD_LAST_REFRESHED:
      return {
        ...state,
        activeDashboard: {
          ...state.activeDashboard,
          refreshed_at: action.payload
        }
      };
    case DASHBOARD_CREATED:
      return {
        ...state,
        dashboards: {
          ...state.dashboards,
          data: [...state.dashboards.data, action.payload]
        }
      };
    case DASHBOARD_DELETED: {
      const newDashboardList = state.dashboards.data.filter(
        (d) => d.id !== action.payload.id
      );
      const newActiveDashboard = newDashboardList[0];
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits },
        dashboards: { ...defaultState.dashboards, data: newDashboardList },
        activeDashboard: newActiveDashboard,
        dashboardDeletionInitiated: false
      };
    }
    case WIDGET_DELETED: {
      const updatedUnitsPosition = { ...state.activeDashboard.units_position };
      _.unset(updatedUnitsPosition, `position.${action.payload}`);
      _.unset(updatedUnitsPosition, `size.${action.payload}`);
      return {
        ...state,
        activeDashboardUnits: {
          ...state.activeDashboardUnits,
          data: state.activeDashboardUnits.data.filter(
            (elem) => elem.id !== action.payload
          )
        },
        activeDashboard: {
          ...state.activeDashboard,
          units_position: updatedUnitsPosition
        }
      };
    }
    case UNITS_ORDER_CHANGED: {
      const activeDashboardIdx = state.dashboards.data.findIndex(
        (elem) => elem.id === state.activeDashboard.id
      );
      return {
        ...state,
        activeDashboardUnits: {
          ...state.activeDashboardUnits,
          data: [...action.payload]
        },
        activeDashboard: {
          ...state.activeDashboard,
          units_position: action.units_position
        },
        dashboards: {
          ...state.dashboards,
          data: [
            ...state.dashboards.data.slice(0, activeDashboardIdx),
            { ...state.activeDashboard, units_position: action.units_position },
            ...state.dashboards.data.slice(activeDashboardIdx + 1)
          ]
        }
      };
    }
    case DASHBOARD_UNMOUNTED:
      return {
        ...state,
        activeDashboardUnits: { ...defaultState.activeDashboardUnits }
      };
    case DASHBOARD_UPDATED: {
      const dashboardIndex = state.dashboards.data.findIndex(
        (dashboard) => dashboard.id === action.payload.id
      );
      const editedDashboard = {
        ...state.dashboards.data[dashboardIndex],
        ...action.payload
      };
      return {
        ...state,
        activeDashboard: editedDashboard,
        dashboards: {
          ...state.dashboards,
          data: [
            ...state.dashboards.data.slice(0, dashboardIndex),
            editedDashboard,
            ...state.dashboards.data.slice(dashboardIndex + 1)
          ]
        }
      };
    }
    case SET_DRAFTS_SELECTED: {
      return {
        ...state,
        draftsSelected: true,
        activeDashboard: defaultState.activeDashboard,
        activeDashboardUnits: defaultState.activeDashboardUnits
      };
    }
    case TOGGLE_DASHBOARD_NEW_FOLDER_MODAL: {
      return {
        ...state,
        showNewFolderModal: action.payload,
        newFolderCreationState: {
          ...apiStates
        }
      };
    }
    case INITIATED_DASHBOARD_NEW_FOLDER_CREATION: {
      return {
        ...state,
        newFolderCreationState: {
          ...apiStates,
          loading: true
        }
      };
    }
    case DASHBOARD_NEW_FOLDER_CREATION_SUCCESSFUL: {
      const { newFolder, dashboardId } = action.payload;
      const currentDashboardFolderId = state.foldersList.data.find((folder) =>
        folder.dashboardIds.find((dId) => dId === dashboardId)
      )?.id;
      const newFolderState = {
        ...newFolder,
        dashboardIds: [dashboardId]
      };
      const updatedFoldersList = state.foldersList.data.map((folder) => {
        if (folder.id === currentDashboardFolderId) {
          return {
            ...folder,
            dashboardIds: folder.dashboardIds.filter(
              (dId) => dId !== dashboardId
            )
          };
        }
        return folder;
      });
      return {
        ...state,
        newFolderCreationState: {
          ...apiStates,
          completed: true
        },
        foldersList: {
          ...state.foldersList,
          data: [newFolderState, ...updatedFoldersList]
        }
      };
    }
    case DASHBOARD_NEW_FOLDER_CREATION_FAILED: {
      return {
        ...state,
        newFolderCreationState: {
          ...apiStates,
          error: true
        }
      };
    }
    case DASHBOARD_FOLDERS_LIST_LOADING: {
      return {
        ...state,
        foldersList: {
          ...defaultState.foldersList,
          loading: true
        }
      };
    }
    case DASHBOARD_FOLDERS_LIST_SUCCESS: {
      const allBoardsFolderId = getAllBoardsFolderId(action.payload);
      if (
        state.dashboards.completed &&
        state.dashboardsArrangementInFoldersCompleted === false
      ) {
        const foldersListWithDashboardIds = getFoldersListWithDashboardIds(
          state.dashboards.data,
          action.payload
        );
        return {
          ...state,
          foldersList: {
            ...defaultState.foldersList,
            completed: true,
            data: foldersListWithDashboardIds
          },
          allBoardsFolderId,
          dashboardsArrangementInFoldersCompleted: true
        };
      }
      return {
        ...state,
        foldersList: {
          ...defaultState.foldersList,
          completed: true,
          data: action.payload.map((folder) => ({
            ...folder,
            dashboardIds: []
          }))
        },
        allBoardsFolderId
      };
    }
    case DASHBOARD_FOLDERS_LIST_ERROR: {
      return {
        ...state,
        foldersList: {
          ...defaultState.foldersList,
          error: true
        }
      };
    }
    case INITIATED_DASHBOARD_MOVE_TO_EXISTING_FOLDER: {
      return {
        ...state,
        addToExistingFolderState: {
          ...apiStates,
          loading: true
        }
      };
    }
    case DASHBOARD_MOVE_TO_EXISTING_FOLDER_SUCCESSFUL: {
      const { dashboardId, folderId } = action.payload;
      const currentDashboardFolderId = state.foldersList.data.find((folder) =>
        folder.dashboardIds.find((dId) => dId === dashboardId)
      )?.id;
      const updatedFoldersList = state.foldersList.data.map((folder) => {
        if (folder.id === currentDashboardFolderId) {
          return {
            ...folder,
            dashboardIds: folder.dashboardIds.filter(
              (dId) => dId !== dashboardId
            )
          };
        }
        if (folder.id === folderId) {
          return {
            ...folder,
            dashboardIds: [...folder.dashboardIds, dashboardId]
          };
        }
        return folder;
      });
      return {
        ...state,
        addToExistingFolderState: {
          ...apiStates,
          completed: true
        },
        foldersList: {
          ...state.foldersList,
          data: updatedFoldersList
        }
      };
    }
    case DASHBOARD_MOVE_TO_EXISTING_FOLDER_FAILED: {
      return {
        ...state,
        addToExistingFolderState: {
          ...apiStates,
          error: true
        }
      };
    }
    case INITIATED_RENAME_DASHBOARD_FOLDER: {
      return {
        ...state,
        renameFolderState: {
          ...apiStates,
          loading: true
        }
      };
    }
    case RENAME_DASHBOARD_FOLDER_SUCCESSFUL: {
      const { folderId, newName } = action.payload;
      const updatedFoldersList = state.foldersList.data.map((folder) => {
        if (folder.id !== folderId) {
          return folder;
        }
        return {
          ...folder,
          name: newName
        };
      });
      return {
        ...state,
        renameFolderState: {
          ...apiStates,
          completed: true
        },
        foldersList: {
          ...state.foldersList,
          data: updatedFoldersList
        }
      };
    }
    case RENAME_DASHBOARD_FOLDER_FAILED: {
      return {
        ...state,
        renameFolderState: {
          ...apiStates,
          error: true
        }
      };
    }
    case INITIATED_DELETE_DASHBOARD_FOLDER: {
      return {
        ...state,
        deleteFolderState: {
          ...apiStates,
          loading: true
        }
      };
    }
    case DELETE_DASHBOARD_FOLDER_SUCCESSFUL: {
      const { folderId } = action.payload;
      const deletedFolderDashboards = state.foldersList.data.find(
        (folder) => folder.id === folderId
      ).dashboardIds;
      const updatedFoldersList = state.foldersList.data.filter(
        (folder) => folder.id !== folderId
      );
      const listWithAdjustedDashboards = updatedFoldersList.map((folder) => {
        if (folder.id === state.allBoardsFolderId) {
          return {
            ...folder,
            dashboardIds: [...deletedFolderDashboards, ...folder.dashboardIds]
          };
        }
        return folder;
      });
      return {
        ...state,
        deleteFolderState: {
          ...apiStates,
          completed: true
        },
        foldersList: {
          ...state.foldersList,
          data: listWithAdjustedDashboards
        }
      };
    }
    case DELETE_DASHBOARD_FOLDER_FAILED: {
      return {
        ...state,
        deleteFolderState: {
          ...apiStates,
          error: true
        }
      };
    }
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };
    case INITIATE_DASHBOARD_DELETION: {
      return {
        ...state,
        dashboardDeletionInitiated: true
      };
    }
    case RESET_DASHBOARD_DELETION_INITIATION: {
      return {
        ...state,
        dashboardDeletionInitiated: false
      };
    }
    case INITIATE_EDIT_DASHBOARD_DETAILS: {
      const { dashboard } = action.payload;
      return {
        ...state,
        editDashboardDetails: {
          initiated: true,
          editDashboard: dashboard
        }
      };
    }
    case ADD_DASHBOARD_MODAL_CLOSE: {
      return {
        ...state,
        editDashboardDetails: {
          ...defaultState.editDashboardDetails
        }
      };
    }
    default:
      return state;
  }
}
