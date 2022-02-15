export const TOGGLE_APIS_CALLED = 'TOGGLE_APIS_CALLED';
export const TOGGLE_MODAL_VISIBILITY = 'TOGGLE_MODAL_VISIBILITY';
export const SET_ACTIVE_ACTION = 'SET_ACTIVE_ACTION';
export const TOGGLE_ADD_TO_DASHBOARD_MODAL = 'TOGGLE_ADD_TO_DASHBOARD_MODAL';

export const ACTION_TYPES = {
  SAVE: 'SAVE_QUERY',
  EDIT: 'EDIT_QUERY',
  ADD_TO_DASHBOARD: 'ADD_TO_DASHBOARD',
};

export const SAVE_QUERY_INITIAL_STATE = {
  apisCalled: false,
  showSaveModal: false,
  showAddToDashModal: false,
  activeAction: null,
};
