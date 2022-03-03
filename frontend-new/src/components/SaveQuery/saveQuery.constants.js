export const TOGGLE_DELETE_MODAL = 'TOGGLE_DELETE_MODAL';
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
  showDeleteModal: false,
};

export const DASHBOARD_PRESENTATION_KEYS = {
  CHART: 'chart',
  TABLE: 'table',
};

export const DEFAULT_DASHBOARD_PRESENTATION = DASHBOARD_PRESENTATION_KEYS.CHART;

export const DASHBOARD_PRESENTATION_LABELS = {
  [DASHBOARD_PRESENTATION_KEYS.CHART]: 'Display Visualisation',
  [DASHBOARD_PRESENTATION_KEYS.TABLE]: 'Display Table',
};
