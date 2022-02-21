import {
  TOGGLE_APIS_CALLED,
  TOGGLE_MODAL_VISIBILITY,
  SET_ACTIVE_ACTION,
  TOGGLE_ADD_TO_DASHBOARD_MODAL,
  TOGGLE_DELETE_MODAL,
} from './saveQuery.constants';

export default function (state, action) {
  const { payload } = action;
  switch (action.type) {
    case TOGGLE_APIS_CALLED: {
      return {
        ...state,
        apisCalled: !state.apisCalled,
      };
    }
    case TOGGLE_MODAL_VISIBILITY: {
      return {
        ...state,
        showSaveModal: !state.showSaveModal,
      };
    }
    case SET_ACTIVE_ACTION: {
      return {
        ...state,
        activeAction: payload,
      };
    }
    case TOGGLE_ADD_TO_DASHBOARD_MODAL: {
      return {
        ...state,
        showAddToDashModal: !state.showAddToDashModal,
      };
    }
    case TOGGLE_DELETE_MODAL: {
      return {
        ...state,
        showDeleteModal: !state.showDeleteModal,
      };
    }
    default:
      return state;
  }
}
