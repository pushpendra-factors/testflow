import {
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN,
  ADD_DASHBOARD_MODAL_OPEN,
  ADD_DASHBOARD_MODAL_CLOSE
} from '../types';
import { defaultState } from './constants';

export default function (state = defaultState, action) {
  switch (action.type) {
    case NEW_DASHBOARD_TEMPLATES_MODAL_OPEN:
      return {
        ...state,
        isNewDashboardTemplateModal: true
      };
    case NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE:
      return {
        ...state,
        isNewDashboardTemplateModal: false
      };
    case ADD_DASHBOARD_MODAL_OPEN:
      return {
        ...state,
        isAddNewDashboardModal: true
      };
    case ADD_DASHBOARD_MODAL_CLOSE:
      return {
        ...state,
        isAddNewDashboardModal: false
      };

    default:
      return state;
  }
}
