import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import { TOTAL_USERS_CRITERIA, EACH_USER_TYPE } from '../../utils/constants';

const defaultState = {
  session_order: {
    between: 0,
    to: 0
  },
  show_criteria: TOTAL_USERS_CRITERIA,
  performance_criteria: EACH_USER_TYPE
};

export const SET_SESSION_ORDER = 'SET_SESSION_ORDER';
export const SET_SHOW_CRITERIA = 'SET_SHOW_CRITERIA';
export const SET_PERFORMANCE_CRITERIA = 'SET_PERFORMANCE_CRITERIA';

export const setSessionOrderAction = (sessionOrder) => {
  return { type: SET_SESSION_ORDER, payload: sessionOrder };
};

export const setShowCriteriaAction = (showCriteria) => {
  return { type: SET_SHOW_CRITERIA, payload: showCriteria };
};

export const setPerformanceCriteriaAction = (performanceCriteria) => {
  return { type: SET_PERFORMANCE_CRITERIA, payload: performanceCriteria };
};

export const setSessionOrder = (sessionOrder) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setSessionOrderAction(sessionOrder)));
    });
  };
};

export const setShowCriteria = (showCriteria) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setShowCriteriaAction(showCriteria)));
    });
  };
};

export const setPerformanceCriteria = (performanceCriteria) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setPerformanceCriteriaAction(performanceCriteria)));
    });
  };
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case SET_SESSION_ORDER:
      return { ...state, session_order: action.payload };
    case SET_SHOW_CRITERIA:
      return {
        ...state,
        show_criteria: action.payload,
        performance_criteria:
          action.payload !== TOTAL_USERS_CRITERIA
            ? EACH_USER_TYPE
            : state.performance_criteria
      };
    case SET_PERFORMANCE_CRITERIA:
      return { ...state, performance_criteria: action.payload };
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };
    default:
      return state;
  }
}
