import {
  BACK_STEP_ONBOARD_FLOW,
  ENABLE_STEP_AND_MOVE_TO_NEXT,
  JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION,
  NEXT_STEP_ONBOARD_FLOW,
  SET_ACTIVE_PROJECT,
  TOGGLE_DISABLED_STATE_NEXT_BUTTON,
  TOGGLE_FACTORS_6SIGNAL_REQUEST,
  TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL,
  UPDATE_ONBOARD_FLOW_STEPS
} from 'Reducers/types';
import { defaultState } from './constants';

export default function (state = defaultState, action) {
  switch (action.type) {
    case UPDATE_ONBOARD_FLOW_STEPS:
      return { ...state, steps: { ...state.steps, ...action.payload } };
    case TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL:
      if (state.isWebsiteVisitorIdentificationVisible === false) {
        return {
          ...state,
          currentStep: 1,
          isWebsiteVisitorIdentificationVisible:
            !state.isWebsiteVisitorIdentificationVisible
        };
      } else {
        return {
          ...state,
          currentStep: null,
          isWebsiteVisitorIdentificationVisible:
            !state.isWebsiteVisitorIdentificationVisible
        };
      }

    case NEXT_STEP_ONBOARD_FLOW:
      return {
        ...state,
        currentStep:
          state.currentStep < 3 ? state.currentStep + 1 : state.currentStep
      };

    case BACK_STEP_ONBOARD_FLOW:
      return {
        ...state,
        currentStep:
          state.currentStep > 1 ? state.currentStep - 1 : state.currentStep
      };
    case TOGGLE_DISABLED_STATE_NEXT_BUTTON:
      return {
        ...state,
        steps: {
          ...state.steps,
          ['step' + action.payload.step]: action.payload.state
        }
      };
    case ENABLE_STEP_AND_MOVE_TO_NEXT:
      return {
        ...state,
        steps: {
          ...state.steps,
          ['step' + action.payload.step]: action.payload.state
        },
        currentStep: action.payload.moveTo
      };
    case TOGGLE_FACTORS_6SIGNAL_REQUEST:
      return {
        ...state,
        factors6SignalKeyRequested: !state.factors6SignalKeyRequested
      };
    case JUMP_TO_STEP_WEBSITE_VISITOR_IDENTIFICATION:
      return { ...state, currentStep: action.payload };
    case SET_ACTIVE_PROJECT: {
      return {
        ...defaultState
      };
    }
    default:
      return state;
  }
}
