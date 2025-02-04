import {
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN,
  ADD_DASHBOARD_MODAL_OPEN,
  ADD_DASHBOARD_MODAL_CLOSE,
  UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE,
  SET_ACTIVE_PROJECT
} from '../types';
import { defaultState } from './constants';

export type DashboardTemplatesControllerType = typeof defaultState;
export default function (
  state: typeof defaultState = defaultState,
  action = { type: '', payload: {} }
) {
  switch (action.type) {
    case NEW_DASHBOARD_TEMPLATES_MODAL_OPEN:
      return {
        ...state,
        isNewDashboardTemplateModal: true
      };
    case NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE:
      return {
        ...state,
        isNewDashboardTemplateModal: false,
        pickedFirstTemplate: null
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
    case UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE:
      return {
        ...state,
        pickedFirstTemplate: action.payload
      };
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };
    default:
      return state;
  }
}
