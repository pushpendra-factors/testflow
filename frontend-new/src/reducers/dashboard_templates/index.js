import {
  TEMPLATES_LOADED,
  TEMPLATES_LOADING,
  TEMPLATES_LOADING_FAILED,
  ACTIVE_TEMPLATE_CHANGE,
  SET_ACTIVE_PROJECT
} from '../types';
import { defaultState } from './constants';

export default function (state = defaultState, action) {
  switch (action.type) {
    case TEMPLATES_LOADING:
      return {
        ...defaultState,
        templates: { ...defaultState.templates, loading: true }
      };
    case TEMPLATES_LOADING_FAILED:
      return {
        ...defaultState,
        templates: { ...defaultState.templates, error: true }
      };
    case TEMPLATES_LOADED:
      return {
        ...defaultState,
        templates: { ...defaultState.templates, data: action.payload }
      };

    case ACTIVE_TEMPLATE_CHANGE:
      return {
        ...state,
        activeTemplate: action.payload
        // activeTemplateUnits: { ...defaultState.activeTemplateUnits },
      };
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };

    default:
      return state;
  }
}
