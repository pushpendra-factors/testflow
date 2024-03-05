import { ACTIVE_DASHBOARD_CHANGE } from 'Reducers/types';
import {
  SET_DRAFTS_SELECTED,
  TOGGLE_DASHBOARD_NEW_FOLDER_MODAL
} from './types';

export const changeActiveDashboardAction = (newActiveDashboard) => ({
  type: ACTIVE_DASHBOARD_CHANGE,
  payload: newActiveDashboard
});

export const makeDraftsActiveAction = () => ({
  type: SET_DRAFTS_SELECTED
});

export const toggleNewFolderModal = (payload) => ({
  type: TOGGLE_DASHBOARD_NEW_FOLDER_MODAL,
  payload
});
